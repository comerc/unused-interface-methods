package test

import (
	"context"
	"io"

	test_data "github.com/comerc/unused-interface-methods/test/data"
)

// MockComplexService создает ComplexService с замоканными интерфейсами
func NewMockComplexService() *test_data.ComplexService {
	return &test_data.ComplexService{
		Reader:      &MockReader{},
		Writer:      &MockWriter{},
		Service:     &MockServiceWithContext{},
		Logger:      &MockLogger{},
		Handler:     &MockEventHandler{},
		Processor:   &MockChannelProcessor{},
		Data:        &MockDataProcessor{},
		Pointer:     &MockPointerHandler{},
		Named:       &MockNamedReturns{},
		ProcessorV1: &MockProcessorV1{},
		ProcessorV2: &MockProcessorV2{},
		Actions:     &MockSimpleActions{},
		Params:      &MockInterfaceParams{},
		Advanced:    &MockAdvancedInterface{},
		AnyProc:     &MockAnyProcessor{},
		AnyHandler:  &MockAnyHandler{},
		StrProc:     &MockStringProcessor{},
		StrHandler:  &MockStringHandler{},
		StrConsumer: &MockStringConsumer{},
		BaseProc:    &MockBaseProcessor{},
		ExtProc:     &MockExtendedProcessor{},
		Direct:      &MockDirectProcessor{},
	}
}

// Моки для каждого интерфейса
type MockReader struct{}

func (m *MockReader) Read(p []byte) (n int, err error) { return len(p), nil }
func (m *MockReader) CustomRead() error                { return nil }

type MockWriter struct{}

func (m *MockWriter) Write(p []byte) (n int, err error) { return len(p), nil }
func (m *MockWriter) CustomWrite() error                { return nil }

type MockServiceWithContext struct{}

func (m *MockServiceWithContext) ProcessWithContext(ctx context.Context, data string) error {
	return nil
}
func (m *MockServiceWithContext) CancelOperation(ctx context.Context) error { return nil }

type MockLogger struct{}

func (m *MockLogger) Log(level string, args ...interface{}) error { return nil }
func (m *MockLogger) Debug(args ...string)                        {}

type MockEventHandler struct{}

func (m *MockEventHandler) OnEvent(callback func(string) error) error           { return nil }
func (m *MockEventHandler) OnError(handler func(error) bool) error              { return nil }
func (m *MockEventHandler) Subscribe(filter func(string) bool, cb func()) error { return nil }

type MockChannelProcessor struct{}

func (m *MockChannelProcessor) SendData(ch chan<- string) error    { return nil }
func (m *MockChannelProcessor) ReceiveData(ch <-chan string) error { return nil }
func (m *MockChannelProcessor) ProcessStream(ch chan string) error { return nil }

type MockDataProcessor struct{}

func (m *MockDataProcessor) ProcessMap(data map[string]interface{}) error { return nil }
func (m *MockDataProcessor) ProcessSlice(data []string) error             { return nil }
func (m *MockDataProcessor) ProcessArray(data [10]int) error              { return nil }

type MockPointerHandler struct{}

func (m *MockPointerHandler) HandlePointer(data *string) error { return nil }
func (m *MockPointerHandler) HandleDoublePtr(data **int) error { return nil }

type MockNamedReturns struct{}

func (m *MockNamedReturns) GetNamedResult() (result string, err error) { return "", nil }
func (m *MockNamedReturns) GetMultipleNamed() (x, y int, success bool, err error) {
	return 0, 0, true, nil
}

type MockProcessorV1 struct{}

func (m *MockProcessorV1) Process(data string) error { return nil }

type MockProcessorV2 struct{}

func (m *MockProcessorV2) Process(data string, options map[string]interface{}) error { return nil }

type MockSimpleActions struct{}

func (m *MockSimpleActions) Start()          {}
func (m *MockSimpleActions) Stop()           {}
func (m *MockSimpleActions) Reset()          {}
func (m *MockSimpleActions) GetStatus() bool { return true }

type MockInterfaceParams struct{}

func (m *MockInterfaceParams) HandleReader(r io.Reader) error        { return nil }
func (m *MockInterfaceParams) HandleWriter(w io.Writer) error        { return nil }
func (m *MockInterfaceParams) HandleBoth(rw io.ReadWriter) error     { return nil }
func (m *MockInterfaceParams) HandleCustom(custom interface{}) error { return nil }

type MockAdvancedInterface struct{}

func (m *MockAdvancedInterface) HandleString(s string) error                  { return nil }
func (m *MockAdvancedInterface) HandleInt(i int) bool                         { return true }
func (m *MockAdvancedInterface) HandleFloat(f float64) string                 { return "" }
func (m *MockAdvancedInterface) HandlePointer(p *string) error                { return nil }
func (m *MockAdvancedInterface) HandleNilPointer() *int                       { return nil }
func (m *MockAdvancedInterface) HandleSlice(slice []string) error             { return nil }
func (m *MockAdvancedInterface) HandleArray(arr [5]int) error                 { return nil }
func (m *MockAdvancedInterface) HandleMap(data map[string]int) error          { return nil }
func (m *MockAdvancedInterface) HandleComplexMap(data map[string][]int) error { return nil }
func (m *MockAdvancedInterface) HandleChannel(ch chan string) error           { return nil }
func (m *MockAdvancedInterface) HandleReadChannel(ch <-chan string) error     { return nil }
func (m *MockAdvancedInterface) HandleWriteChannel(ch chan<- string) error    { return nil }
func (m *MockAdvancedInterface) HandleFunc(f func(string) error) error        { return nil }
func (m *MockAdvancedInterface) HandleComplexFunc(f func(int, string) (bool, error)) error {
	return nil
}
func (m *MockAdvancedInterface) HandleInterface(i interface{}) error                  { return nil }
func (m *MockAdvancedInterface) HandleContext(ctx context.Context) error              { return nil }
func (m *MockAdvancedInterface) HandleVariadic(args ...string) error                  { return nil }
func (m *MockAdvancedInterface) HandleMixedVariadic(prefix string, args ...int) error { return nil }
func (m *MockAdvancedInterface) GetResult() bool                                      { return true }
func (m *MockAdvancedInterface) Clear()                                               {}
func (m *MockAdvancedInterface) GetMultiple() (string, int, error)                    { return "", 0, nil }
func (m *MockAdvancedInterface) GetNamedReturns() (result string, success bool)       { return "", true }

type MockAnyProcessor struct{}

func (m *MockAnyProcessor) Process(data interface{}) error { return nil }

type MockAnyHandler struct{}

func (m *MockAnyHandler) Process(data interface{}, options interface{}) error { return nil }

type MockStringProcessor struct{}

func (m *MockStringProcessor) Process(data string) error { return nil }

type MockStringHandler struct{}

func (m *MockStringHandler) Process(data string) error { return nil }

type MockStringConsumer struct{}

func (m *MockStringConsumer) Process(data string) error { return nil }

type MockBaseProcessor struct{}

func (m *MockBaseProcessor) Process() error { return nil }

type MockExtendedProcessor struct{}

func (m *MockExtendedProcessor) Process() error { return nil }
func (m *MockExtendedProcessor) Extra() bool    { return true }

type MockDirectProcessor struct{}

func (m *MockDirectProcessor) ProcessDirect(data string) error { return nil }
