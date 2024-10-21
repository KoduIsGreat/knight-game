package nw

//go:generate go-enum -f=$GOFILE --noprefix --marshal --nocase -t values.tmpl
/*
ENUM(
MsgAuth
MsgAuthAck
MsgConnect
MsgDisconnect
MsgLobbyCreate
MsgLobbyCreated
MsgLobbyDeleted
MsgLobbyGameStart
MsgLobbyClientReady
MsgLobbyClientJoin
MsgLobbyClientLeave
MsgClientInput
MsgServerState
)
*/
type MessageHeader uint8
