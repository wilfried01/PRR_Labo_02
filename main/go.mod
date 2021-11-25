module PRR_Labo_02

go 1.17

require (
	server v1.0.0
	configuration v1.0.0
)

replace (
	server v1.0.0 => ../server
	configuration v1.0.0 => ../configuration
)
