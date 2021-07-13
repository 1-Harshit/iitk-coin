package config

// Path of DB
const DBpath string = "./data.db"

// Max coins one can have
const MaxCoins int = 10000

// Minimum Events needed for transfer
const MinEvents int = 5

// tax
const IntraBatchTax float64 = 0.02
const InterBatchTax float64 = 0.33

// If someone can reedeem now
const IsStoreOpen bool = true