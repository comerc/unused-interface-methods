package test

import (
	"context"
	"testing"

	test_data "github.com/comerc/unused-interface-methods/test/data"
)

// ===============================
// СПЕЦИАЛЬНЫЕ КЕЙСЫ 17-22
// ===============================

// Кейс 17: Использование через type assertion
func TestCase17_TypeAssertion(t *testing.T) {
	t.Run("Reader.CustomRead через type assertion", func(t *testing.T) {
		var cs test_data.ComplexService
		var iface interface{} = cs.Reader
		if r, ok := iface.(test_data.Reader); ok {
			r.CustomRead()
		}
	})

	t.Run("AdvancedInterface методы через type assertion", func(t *testing.T) {
		var cs test_data.ComplexService
		if adv, ok := interface{}(cs.Advanced).(test_data.AdvancedInterface); ok {
			adv.HandleInt(42)
			adv.HandleFloat(3.14)
		}
	})
}

// Кейс 18: Использование методов в горутинах
func TestCase18_GoroutineUsage(t *testing.T) {
	t.Run("Stop и HandleNilPointer в горутине", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.GoroutineUsage()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})
}

// Кейс 19: Использование методов как значений
func TestCase19_MethodValues(t *testing.T) {
	t.Run("Методы как значения", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.MethodValues()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})
}

// Кейс 20: Использование в switch/type switch
func TestCase20_SwitchUsage(t *testing.T) {
	t.Run("Методы в type switch", func(t *testing.T) {
		var cs test_data.ComplexService
		var value interface{} = cs.Advanced

		switch v := value.(type) {
		case test_data.AdvancedInterface:
			v.HandleReadChannel(make(<-chan string))
			v.HandleWriteChannel(make(chan<- string))
			v.HandleComplexFunc(func(int, string) (bool, error) { return true, nil })
		}

		value = cs.Service
		switch v := value.(type) {
		case test_data.ServiceWithContext:
			v.CancelOperation(context.Background())
		}
	})
}

// Кейс 21: Использование в defer
func TestCase21_DeferUsage(t *testing.T) {
	t.Run("Методы в defer", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DeferUsage()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})
}

// Кейс 22: Использование методов через рефлексию
func TestCase22_ReflectionUsage(t *testing.T) {
	t.Run("Методы через рефлексию", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.ReflectionUsage()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})

	t.Run("Методы через interface{}", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.TypeAssertions()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})

	t.Run("Методы с interface{} параметрами", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.TypeAssertions()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})

	t.Run("Методы через интерфейсные переменные", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.TypeAssertions()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})

	t.Run("Методы через разные интерфейсы", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.TypeAssertions()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})
}
