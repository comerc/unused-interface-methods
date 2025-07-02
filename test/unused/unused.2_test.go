package unused

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCase17_TypeAssertion проверяет использование через type assertion
func TestCase17_TypeAssertion(t *testing.T) {
	t.Run("Reader.CustomRead не должен находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: "+pkgPath+".Reader.CustomRead",
			"метод CustomRead не должен определяться в stage1, т.к. используется через type assertion")
	})

	t.Run("AdvancedInterface методы не должны находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleInt",
			"метод HandleInt не должен определяться в stage1, т.к. используется через type assertion")
		assert.NotContains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleFloat",
			"метод HandleFloat не должен определяться в stage1, т.к. используется через type assertion")
	})
}

// TestCase18_Goroutines проверяет использование в горутинах
func TestCase18_Goroutines(t *testing.T) {
	t.Run("SimpleActions.Stop не должен находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: SimpleActions.Stop",
			"метод Stop не должен определяться в stage1, т.к. используется в горутине")
	})

	t.Run("ChannelProcessor методы не должны находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: ChannelProcessor.HandleReadChannel",
			"метод HandleReadChannel не должен определяться в stage1, т.к. используется в горутине")
		assert.NotContains(t, testOutput, "OK: ChannelProcessor.HandleWriteChannel",
			"метод HandleWriteChannel не должен определяться в stage1, т.к. используется в горутине")
	})
}

// TestCase19_MethodValues проверяет использование method values
func TestCase19_MethodValues(t *testing.T) {
	t.Run("Logger.Debug не должен находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: Logger.Debug",
			"метод Debug не должен определяться в stage1, т.к. используется как значение функции")
	})

	t.Run("AdvancedInterface методы не должны находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: AdvancedInterface.HandleArray",
			"метод HandleArray не должен определяться в stage1, т.к. используется как значение функции")
		assert.NotContains(t, testOutput, "OK: AdvancedInterface.HandleComplexMap",
			"метод HandleComplexMap не должен определяться в stage1, т.к. используется как значение функции")
	})
}

// TestCase20_SwitchUsage проверяет использование в type switch
func TestCase20_SwitchUsage(t *testing.T) {
	t.Run("ServiceWithContext.CancelOperation не должен находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: ServiceWithContext.CancelOperation",
			"метод CancelOperation не должен определяться в stage1, т.к. используется в type switch")
	})

	t.Run("ChannelProcessor методы не должны находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: ChannelProcessor.HandleReadChannel",
			"метод HandleReadChannel не должен определяться в stage1, т.к. используется в type switch")
		assert.NotContains(t, testOutput, "OK: ChannelProcessor.HandleWriteChannel",
			"метод HandleWriteChannel не должен определяться в stage1, т.к. используется в type switch")
		assert.NotContains(t, testOutput, "OK: ChannelProcessor.HandleComplexFunc",
			"метод HandleComplexFunc не должен определяться в stage1, т.к. используется в type switch")
	})
}

// TestCase21_DeferUsage проверяет использование в defer
func TestCase21_DeferUsage(t *testing.T) {
	t.Run("SimpleActions.Reset не должен находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: SimpleActions.Reset",
			"метод Reset не должен определяться в stage1, т.к. используется в defer")
	})

	t.Run("ChannelProcessor методы не должны находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: ChannelProcessor.HandleReadChannel",
			"метод HandleReadChannel не должен определяться в stage1, т.к. используется в defer")
		assert.NotContains(t, testOutput, "OK: ChannelProcessor.HandleWriteChannel",
			"метод HandleWriteChannel не должен определяться в stage1, т.к. используется в defer")
	})
}

// TestCase22_ReflectionUsage проверяет использование через reflection
func TestCase22_ReflectionUsage(t *testing.T) {
	t.Run("SimpleActions.Reset не должен находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: SimpleActions.Reset",
			"метод Reset не должен определяться в stage1, т.к. используется через reflection")
	})

	t.Run("ChannelProcessor методы не должны находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: ChannelProcessor.HandleReadChannel",
			"метод HandleReadChannel не должен определяться в stage1, т.к. используется через reflection")
		assert.NotContains(t, testOutput, "OK: ChannelProcessor.HandleWriteChannel",
			"метод HandleWriteChannel не должен определяться в stage1, т.к. используется через reflection")
	})
}
