package main

var (
	logger *Logger
	config *Config
)

func main() {
	config = GetConfig()
	logger = NewLogger(config.LogConfig)
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warning("Warning message")
	logger.Error("Error message")
	logger.Critical("Critical message")
	logger.Print("Print message")
	logger.Printf("Printf message: %s", "Yopta")
	logger.Fatalf("Fatal message: %s", "Nah")
}
