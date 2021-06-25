package config

// Path of DB
var DBpath string = "./data.db"

// Max coins one can have
var MaxCoins int = 10000

// Minimum Events needed for transfer
var MinEvents int = 5

// tax
var IntraBatchTax float64 = 0.02
var InterBatchTax float64 = 0.33