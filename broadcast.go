package socketio

import "sync"

// BroadcastConn is overall intarface same a intarface about connection Conn in go-socket.io
type BroadcastConn interface {
	ID() string                        // ID returns session id by connect
	Emit(msg string, v ...interface{}) // Emit message to client
}

// Broadcast is the adaptor to handle broadcasts & rooms for socket.io server API
type Broadcast interface {
	Join(room string, connection BroadcastConn)   // Join causes the connection to join a room
	Leave(room string, connection BroadcastConn)  // Leave causes the connection to leave a room
	LeaveAll(connection BroadcastConn)            // LeaveAll causes given connection to leave all rooms
	Clear(room string)                            // Clear causes removal of all connections from the room
	Send(room, event string, args ...interface{}) // Send will send an event with args to the room
	SendAll(event string, args ...interface{})    // SendAll will send an event with args to all the rooms
	Len(room string) int                          // Len gives number of connections in the room
	Rooms(connection BroadcastConn) []string      // Gives list of all the rooms if no connection given, else list of all the rooms the connection joined
}

// broadcast gives Join, Leave & BroadcastTO server API support to socket.io along with room management
type broadcast struct {
	lock  sync.RWMutex                        // access lock for rooms
	rooms map[string]map[string]BroadcastConn // map of rooms where each room contains a map of connection id to connections in that room
}

// NewBroadcast creates a new broadcast adapter
func NewBroadcast() Broadcast {
	return &broadcast{rooms: make(map[string]map[string]BroadcastConn)}
}

// Join joins the given connection to the broadcast room
func (broadcast *broadcast) Join(room string, connection BroadcastConn) {
	broadcast.lock.Lock()

	// check if room already has connection mappings, if not then create one
	if _, ok := broadcast.rooms[room]; !ok {
		broadcast.rooms[room] = make(map[string]BroadcastConn)
	}

	// add the connection to the rooms connection map
	broadcast.rooms[room][connection.ID()] = connection

	broadcast.lock.Unlock()
}

// Leave leaves the given connection from given room (if exist)
func (broadcast *broadcast) Leave(room string, connection BroadcastConn) {
	broadcast.lock.Lock()

	// check if rooms connection
	if connections, ok := broadcast.rooms[room]; ok {
		// remove the connection from the room
		delete(connections, connection.ID())

		// check if no more connection is left to the room, then delete the room
		if len(connections) == 0 {
			delete(broadcast.rooms, room)
		}
	}

	broadcast.lock.Unlock()
}

// LeaveAll leaves the given connection from all rooms
func (broadcast *broadcast) LeaveAll(connection BroadcastConn) {
	broadcast.lock.Lock()

	// iterate through each room
	for room, connections := range broadcast.rooms {
		// remove the connection from the rooms connections
		delete(connections, connection.ID())

		// check if no more connection is left to the room, then delete the room
		if len(connections) == 0 {
			delete(broadcast.rooms, room)
		}
	}
	broadcast.lock.Unlock()
}

// Clear clears the room
func (broadcast *broadcast) Clear(room string) {
	broadcast.lock.Lock()

	delete(broadcast.rooms, room)

	broadcast.lock.Unlock()
}

// Send sends given event & args to all the connections in the specified room
func (broadcast *broadcast) Send(room, event string, args ...interface{}) {
	broadcast.lock.RLock()

	// iterate through each connection in the room
	for _, connection := range broadcast.rooms[room] {
		// emit the event to the connection
		connection.Emit(event, args...)
	}
	broadcast.lock.RUnlock()
}

// SendAll sends given event & args to all the connections to all the rooms
func (broadcast *broadcast) SendAll(event string, args ...interface{}) {
	broadcast.lock.RLock()

	// iterate through each room
	for _, connections := range broadcast.rooms {
		// iterate through each connection in the room
		for _, connection := range connections {
			// emit the event to the connection
			connection.Emit(event, args...)
		}
	}
	broadcast.lock.RUnlock()
}

// Len gives number of connections in the room
func (broadcast *broadcast) Len(room string) int {
	broadcast.lock.RLock()
	conn2Room := len(broadcast.rooms[room])
	broadcast.lock.RUnlock()
	return conn2Room
}

// Rooms gives the list of all the rooms available for broadcast in case of
// no connection is given, in case of a connection is given, it gives
// list of all the rooms the connection is joined to
func (broadcast *broadcast) Rooms(connection BroadcastConn) []string {
	broadcast.lock.RLock()

	// new list of all the room names
	var rooms []string

	if connection == nil {
		rooms = make([]string, 0, len(broadcast.rooms))
		// iterate through the rooms map and add the room name to the above list
		for room := range broadcast.rooms {
			rooms = append(rooms, room)
		}
	} else {
		rooms = make([]string, 0)
		// iterate through the rooms map and add the room name to the above list
		for room, connections := range broadcast.rooms {
			// check if the connection is joined to the room
			if _, ok := connections[connection.ID()]; ok {
				// add the room name to the list
				rooms = append(rooms, room)
			}
		}
	}
	broadcast.lock.RUnlock()

	return rooms
}
