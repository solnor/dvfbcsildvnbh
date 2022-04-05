module TTK4145-Project

go 1.17

require driver v0.0.0-00010101000000-000000000000

require (
	backup v0.0.0-00010101000000-000000000000 // indirect
	elevator v0.0.0-00010101000000-000000000000 // indirect
	network v0.0.0-00010101000000-000000000000 // indirect
)

replace (
	backup => ./backup
	driver => ./driver
	elevator => ./elevator
	network => ./network
)
