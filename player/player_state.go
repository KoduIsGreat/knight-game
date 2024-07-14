package player

//go:generate go-enum -f=$GOFILE --marshal  -t values.tmpl

/*
ENUM(
IDLE
RUNNING
JUMPING
FALLING
)
*/
type PlayerState int
