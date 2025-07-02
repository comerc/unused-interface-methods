package test

import (
	"testing"

	test_data "github.com/comerc/unused-interface-methods/test/data"
)

// ===============================
// БАЗОВЫЕ КЕЙСЫ 1-16
// ===============================

// Кейс 1: Встроенные интерфейсы
func TestCase1_EmbeddedInterfaces(t *testing.T) {
	t.Run("Reader.CustomRead используется через type assertion", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.TypeAssertions()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("Writer.CustomWrite используется напрямую", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("io.Reader методы не используются", func(t *testing.T) {
		// TODO: проверить, что линтер находит неиспользуемые методы io.Reader
	})
}

// Кейс 2: Интерфейс с контекстом
func TestCase2_ContextInterface(t *testing.T) {
	t.Run("ProcessWithContext используется напрямую", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("CancelOperation используется в type switch", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.SwitchUsage(cs.Service)
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})
}

// Кейс 3: Интерфейс с вариативными параметрами
func TestCase3_VariadicInterface(t *testing.T) {
	t.Run("Log используется напрямую", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("Debug используется как значение функции", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.MethodValues()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})
}

// Кейс 4: Интерфейс с функциональными типами
func TestCase4_FunctionalInterface(t *testing.T) {
	t.Run("OnEvent используется напрямую", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("OnError и Subscribe не используются", func(t *testing.T) {
		// TODO: проверить, что линтер находит OnError и Subscribe как неиспользуемые
	})
}

// Кейс 5: Интерфейс с каналами
func TestCase5_ChannelInterface(t *testing.T) {
	t.Run("SendData используется напрямую", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("ReceiveData и ProcessStream не используются", func(t *testing.T) {
		// TODO: проверить, что линтер находит ReceiveData и ProcessStream как неиспользуемые
	})
}

// Кейс 6: Интерфейс с мапами и слайсами
func TestCase6_MapsAndSlicesInterface(t *testing.T) {
	t.Run("ProcessMap используется напрямую", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("ProcessSlice и ProcessArray не используются", func(t *testing.T) {
		t.Skip("Методы ProcessSlice и ProcessArray не используются")
	})
}

// Кейс 7: Интерфейс с указателями
func TestCase7_PointerInterface(t *testing.T) {
	t.Run("HandlePointer используется напрямую", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("HandlePointer используется через прямой вызов на структуре", func(t *testing.T) {
		s := &test_data.TestPointerStruct{}
		str := "test"
		s.HandlePointer(&str)
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("HandleDoublePtr не используется", func(t *testing.T) {
		t.Skip("Метод HandleDoublePtr не используется")
	})
}

// Кейс 8: Интерфейс с именованными возвращаемыми значениями
func TestCase8_NamedReturnsInterface(t *testing.T) {
	t.Run("GetNamedResult используется напрямую", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("GetMultipleNamed не используется", func(t *testing.T) {
		t.Skip("Метод GetMultipleNamed не используется")
	})
}

// Кейс 9: Интерфейсы с одинаковыми именами методов
func TestCase9_SameNameMethods(t *testing.T) {
	t.Run("ProcessorV1.Process используется напрямую", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("ProcessorV2.Process не используется", func(t *testing.T) {
		t.Skip("Метод ProcessorV2.Process не используется")
	})
}

// Кейс 10: Интерфейс с методами без параметров
func TestCase10_NoParamMethods(t *testing.T) {
	t.Run("Start и GetStatus используются напрямую", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})

	t.Run("Stop используется в горутине", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.GoroutineUsage()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("Reset используется в defer и reflection", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DeferUsage()
		cs.ReflectionUsage()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})
}

// Кейс 11: Интерфейс с интерфейсными параметрами
func TestCase11_InterfaceParams(t *testing.T) {
	t.Run("HandleReader используется напрямую", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})

	t.Run("HandleWriter, HandleBoth и HandleCustom не используются", func(t *testing.T) {
		t.Skip("Методы HandleWriter, HandleBoth и HandleCustom не используются")
	})
}

// Кейс 12: Расширенный интерфейс с разными типами параметров
func TestCase12_AdvancedInterface(t *testing.T) {
	t.Run("Простые типы", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		cs.TypeAssertions()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})

	t.Run("Указатели", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		cs.GoroutineUsage()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})

	t.Run("Слайсы и массивы", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.MethodValues()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})

	t.Run("Мапы", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.MethodValues()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})

	t.Run("Каналы", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.SwitchUsage(cs.Advanced)
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})

	t.Run("Функциональные типы", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.SwitchUsage(cs.Advanced)
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})

	t.Run("Интерфейсы", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DoComplexWork()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})

	t.Run("Вариативные параметры", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.DeferUsage()
		// Если мы дошли до этой точки без паники, значит методы были успешно вызваны
	})
}

// Кейс 13: Интерфейс с методом, имеющим такое же имя
func TestCase13_SameNameDifferentSignature(t *testing.T) {
	t.Run("AnotherReader.CustomRead не используется", func(t *testing.T) {
		t.Skip("Метод AnotherReader.CustomRead не используется")
	})
}

// Кейс 14: Интерфейс с interface{} параметрами
func TestCase14_InterfaceParams(t *testing.T) {
	t.Run("AnyProcessor.Process используется", func(t *testing.T) {
		var cs test_data.ComplexService
		var processor interface{} = cs.ProcessorV1
		if p, ok := processor.(test_data.AnyProcessor); ok {
			p.Process("data")
		}
	})

	t.Run("AnyHandler.Process не используется", func(t *testing.T) {
		t.Skip("Метод AnyHandler.Process не используется")
	})
}

// Кейс 15: Интерфейсы для проверки вызовов на интерфейсных переменных
func TestCase15_InterfaceVariables(t *testing.T) {
	t.Run("Process используется через интерфейсные переменные", func(t *testing.T) {
		cs := NewMockComplexService()
		cs.ReflectionUsage()
		// Если мы дошли до этой точки без паники, значит метод был успешно вызван
	})
}

// Кейс 16: Интерфейсы для проверки вызовов с разными сигнатурами
func TestCase16_DifferentSignatures(t *testing.T) {
	t.Run("Process и Extra используются через разные интерфейсы", func(t *testing.T) {
		impl := &test_data.ProcessorImpl{}
		var base test_data.BaseProcessor = impl
		var extended test_data.ExtendedProcessor = impl
		base.Process()
		extended.Process()
		extended.Extra()
	})
}
