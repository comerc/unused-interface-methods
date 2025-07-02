package unused

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const pkgPath = "github.com/comerc/unused-interface-methods/test/data"

// TestCase1_EmbeddedInterfaces проверяет обработку встроенных интерфейсов
func TestCase1_EmbeddedInterfaces(t *testing.T) {
	t.Run("Reader.CustomRead используется через type assertion", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".Reader.CustomRead",
			"метод CustomRead должен определяться в stage1, т.к. это прямой вызов метода")
	})

	t.Run("Writer.CustomWrite используется напрямую", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".Writer.CustomWrite",
			"метод CustomWrite должен определяться как используемый напрямую в stage1")
	})
}

// TestCase2_ContextInterface проверяет интерфейс с контекстом
func TestCase2_ContextInterface(t *testing.T) {
	t.Run("ProcessWithContext используется напрямую", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".ServiceWithContext.ProcessWithContext",
			"метод ProcessWithContext должен определяться как используемый напрямую в stage1")
	})

	t.Run("CancelOperation используется напрямую в type switch", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".ServiceWithContext.CancelOperation",
			"метод CancelOperation должен определяться как используемый напрямую в stage1, т.к. это тоже прямой вызов o.i.Method()")
	})
}

// TestCase3_VariadicInterface проверяет интерфейс с вариативными параметрами
func TestCase3_VariadicInterface(t *testing.T) {
	t.Run("Log используется напрямую", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".Logger.Log",
			"метод Log должен определяться как используемый напрямую в stage1")
	})

	t.Run("Debug не должен находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: "+pkgPath+".Logger.Debug",
			"метод Debug не должен определяться в stage1, т.к. используется как значение функции")
	})
}

// TestCase4_FunctionalInterface проверяет интерфейс с функциональными типами
func TestCase4_FunctionalInterface(t *testing.T) {
	t.Run("OnEvent используется напрямую", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".EventHandler.OnEvent",
			"метод OnEvent должен определяться как используемый напрямую в stage1")
	})

	t.Run("OnError и Subscribe не используются", func(t *testing.T) {
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".EventHandler.OnError",
			"метод OnError должен быть помечен как неиспользуемый в stage2")
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".EventHandler.Subscribe",
			"метод Subscribe должен быть помечен как неиспользуемый в stage2")
	})
}

// TestCase5_ChannelInterface проверяет интерфейс с каналами
func TestCase5_ChannelInterface(t *testing.T) {
	t.Run("SendData используется напрямую", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".ChannelProcessor.SendData",
			"метод SendData должен определяться как используемый напрямую в stage1")
	})

	t.Run("ReceiveData и ProcessStream не используются", func(t *testing.T) {
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".ChannelProcessor.ReceiveData",
			"метод ReceiveData должен быть помечен как неиспользуемый в stage2")
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".ChannelProcessor.ProcessStream",
			"метод ProcessStream должен быть помечен как неиспользуемый в stage2")
	})
}

// TestCase6_DataProcessor проверяет интерфейс с мапами и слайсами
func TestCase6_DataProcessor(t *testing.T) {
	t.Run("ProcessMap используется напрямую", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".DataProcessor.ProcessMap",
			"метод ProcessMap должен определяться как используемый напрямую в stage1")
	})

	t.Run("ProcessSlice и ProcessArray не используются", func(t *testing.T) {
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".DataProcessor.ProcessSlice",
			"метод ProcessSlice должен быть помечен как неиспользуемый в stage2")
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".DataProcessor.ProcessArray",
			"метод ProcessArray должен быть помечен как неиспользуемый в stage2")
	})
}

// TestCase7_PointerHandler проверяет интерфейс с указателями
func TestCase7_PointerHandler(t *testing.T) {
	t.Run("HandlePointer используется напрямую", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".PointerHandler.HandlePointer",
			"метод HandlePointer должен определяться как используемый напрямую в stage1")
	})

	t.Run("HandleDoublePtr не используется", func(t *testing.T) {
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".PointerHandler.HandleDoublePtr",
			"метод HandleDoublePtr должен быть помечен как неиспользуемый в stage2")
	})
}

// TestCase8_NamedReturns проверяет интерфейс с именованными возвращаемыми значениями
func TestCase8_NamedReturns(t *testing.T) {
	t.Run("GetNamedResult используется напрямую", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".NamedReturns.GetNamedResult",
			"метод GetNamedResult должен определяться как используемый напрямую в stage1")
	})

	t.Run("GetMultipleNamed не используется", func(t *testing.T) {
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".NamedReturns.GetMultipleNamed",
			"метод GetMultipleNamed должен быть помечен как неиспользуемый в stage2")
	})
}

// TestCase9_ProcessorVersions проверяет интерфейсы с одинаковыми именами методов
func TestCase9_ProcessorVersions(t *testing.T) {
	t.Run("ProcessorV1.Process используется напрямую", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".ProcessorV1.Process",
			"метод Process в ProcessorV1 должен определяться как используемый напрямую в stage1")
	})

	t.Run("ProcessorV2.Process не используется", func(t *testing.T) {
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".ProcessorV2.Process",
			"метод Process в ProcessorV2 должен быть помечен как неиспользуемый в stage2")
	})
}

// TestCase10_SimpleActions проверяет интерфейс с методами без параметров
func TestCase10_SimpleActions(t *testing.T) {
	t.Run("Start и GetStatus используются напрямую", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".SimpleActions.Start",
			"метод Start должен определяться как используемый напрямую в stage1")
		assert.Contains(t, testOutput, "OK: "+pkgPath+".SimpleActions.GetStatus",
			"метод GetStatus должен определяться как используемый напрямую в stage1")
	})

	t.Run("Stop и Reset не должны находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: "+pkgPath+".SimpleActions.Stop",
			"метод Stop не должен определяться в stage1, т.к. используется в горутине")
		assert.NotContains(t, testOutput, "OK: "+pkgPath+".SimpleActions.Reset",
			"метод Reset не должен определяться в stage1, т.к. используется в defer и reflection")
	})
}

// TestCase11_InterfaceParams проверяет интерфейс с интерфейсными параметрами
func TestCase11_InterfaceParams(t *testing.T) {
	t.Run("HandleReader используется напрямую", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".InterfaceParams.HandleReader",
			"метод HandleReader должен определяться как используемый напрямую в stage1")
	})

	t.Run("HandleWriter, HandleBoth и HandleCustom не используются", func(t *testing.T) {
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".InterfaceParams.HandleWriter",
			"метод HandleWriter должен быть помечен как неиспользуемый в stage2")
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".InterfaceParams.HandleBoth",
			"метод HandleBoth должен быть помечен как неиспользуемый в stage2")
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".InterfaceParams.HandleCustom",
			"метод HandleCustom должен быть помечен как неиспользуемый в stage2")
	})
}

// TestCase12_AdvancedInterface проверяет расширенный интерфейс с разными типами параметров
func TestCase12_AdvancedInterface(t *testing.T) {
	// Все эти методы используются напрямую в DoComplexWork
	t.Run("Методы используемые напрямую", func(t *testing.T) {
		assert.Contains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleString")
		assert.Contains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandlePointer")
		assert.Contains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleSlice")
		assert.Contains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleMap")
		assert.Contains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleChannel")
		assert.Contains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleFunc")
		assert.Contains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleInterface")
		assert.Contains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleContext")
		assert.Contains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleVariadic")
		assert.Contains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.GetResult")
		assert.Contains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.GetMultiple")
	})

	// Эти методы используются другими способами
	t.Run("Методы используемые не напрямую", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleInt",
			"метод HandleInt не должен определяться в stage1, т.к. используется через type assertion")
		assert.NotContains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleFloat",
			"метод HandleFloat не должен определяться в stage1, т.к. используется через type assertion")
		assert.NotContains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleArray",
			"метод HandleArray не должен определяться в stage1, т.к. используется как значение функции")
		assert.NotContains(t, testOutput, "OK: "+pkgPath+".AdvancedInterface.HandleComplexMap",
			"метод HandleComplexMap не должен определяться в stage1, т.к. используется как значение функции")
	})
}

// TestCase13_SameNameDifferentSignature проверяет методы с одинаковыми именами
func TestCase13_SameNameDifferentSignature(t *testing.T) {
	t.Run("Reader.CustomRead не должен находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: "+pkgPath+".Reader.CustomRead",
			"метод CustomRead не должен определяться в stage1, т.к. используется через type assertion")
	})

	t.Run("AnotherReader.CustomRead не используется", func(t *testing.T) {
		assert.Contains(t, testOutput, "UNUSED: "+pkgPath+".AnotherReader.CustomRead",
			"метод CustomRead должен быть помечен как неиспользуемый в stage2")
	})
}

// TestCase14_AnyProcessor проверяет интерфейс с interface{} параметрами
func TestCase14_AnyProcessor(t *testing.T) {
	t.Run("AnyProcessor.Process не должен находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: AnyProcessor.Process",
			"метод Process не должен определяться в stage1, т.к. используется через interface{}")
	})

	t.Run("AnyHandler.Process не используется", func(t *testing.T) {
		assert.Contains(t, testOutput, "UNUSED: AnyHandler.Process",
			"метод Process должен быть помечен как неиспользуемый в stage2")
	})
}

// TestCase15_InterfaceVariables проверяет использование через интерфейсные переменные
func TestCase15_InterfaceVariables(t *testing.T) {
	t.Run("StringProcessor и StringConsumer не должны находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: StringProcessor.Process",
			"метод Process не должен определяться в stage1, т.к. используется через интерфейсные переменные")
		assert.NotContains(t, testOutput, "OK: StringConsumer.Process",
			"метод Process не должен определяться в stage1, т.к. используется через интерфейсные переменные")
	})

	t.Run("StringHandler не используется", func(t *testing.T) {
		assert.Contains(t, testOutput, "UNUSED: StringHandler.Process",
			"метод Process должен быть помечен как неиспользуемый в stage2")
	})
}

// TestCase16_DifferentSignatures проверяет использование через разные интерфейсы
func TestCase16_DifferentSignatures(t *testing.T) {
	t.Run("BaseProcessor и ExtendedProcessor не должны находиться в stage1", func(t *testing.T) {
		assert.NotContains(t, testOutput, "OK: BaseProcessor.Process",
			"метод Process не должен определяться в stage1, т.к. используется через разные интерфейсы")
		assert.NotContains(t, testOutput, "OK: ExtendedProcessor.Process",
			"метод Process не должен определяться в stage1, т.к. используется через разные интерфейсы")
		assert.NotContains(t, testOutput, "OK: ExtendedProcessor.Extra",
			"метод Extra не должен определяться в stage1, т.к. используется через разные интерфейсы")
	})
}
