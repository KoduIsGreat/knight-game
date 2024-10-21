package nw

//go:generate go-enum -f=$GOFILE --noprefix --marshal --nocase -t values.tmpl

/*
ENUM(
FmtText
FmtJSON
FmtBinary
)
*/
type MessageFmt uint8
