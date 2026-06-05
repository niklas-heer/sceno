package testutil

// ValidMinimalKDL is the smallest buildable diagram.
const ValidMinimalKDL = `diagram layout=auto gap=32 {
  shape box a "A" at=0,0
  shape box b "B" at=1,0
  edge a -> b
}`

// ValidVerticalKDL stacks nodes in one column.
const ValidVerticalKDL = `diagram layout=auto gap=36 {
  shape box top "Top" icon=zap iconPos=top at=0,0
  shape box mid "Mid" icon=queue iconPos=top at=0,1
  shape box bot "Bot" icon=storage iconPos=top at=0,2
  edge top -> mid
  edge mid -> bot
}`

// ValidSeeds are corpus entries for fuzz tests (FastCheck-style shrinking seeds).
var ValidSeeds = []string{
	ValidMinimalKDL,
	ValidVerticalKDL,
	`diagram title="T" layout=auto gap=40 style=polished {
  shape actor u "User" at=0,0
  shape box api "API" icon=api at=1,0
  edge u -> api fromSide=right toSide=left
}`,
	`diagram layout=free gap=24 {
  shape box a "A" x=10 y=20
  shape box b "B" x=200 y=20
  edge a -> b
}`,
}
