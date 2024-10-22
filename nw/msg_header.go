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
MsgLobbyGameStarted
MsgLobbyClientsNotReady
MsgLobbyClientReady
MsgLobbyClientJoin
MsgLobbyClientLeave
MsgLobbiesSync
MsgLobbiesSynced
MsgLobbyPromote
MsgLobbyPromoted
MsgLobbyKick
MsgLobbyKicked
MsgClientInput
MsgServerState
)
*/
type MessageHeader uint8
