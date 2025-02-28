package bilogs

type LogBuffer struct {
	logs chan string
}

func NewLogBuffer(bufferSize int) *LogBuffer {
	return &LogBuffer{
		logs: make(chan string, bufferSize),
	}
}

func (b *LogBuffer) AddLog(log string) {
	select {
	case b.logs <- log:
	default:
	}
}

func (b *LogBuffer) GetLogs() <-chan string {
	return b.logs
}

func (b *LogBuffer) Close() {
	close(b.logs)
}