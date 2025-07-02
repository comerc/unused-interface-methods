# Техническое задание (черновик)

Базовый кейс, который должен обрабатывать stage1

```go
package some_name

// частично применяемый интерфейс
type Interface interface {
  SomeMethod()
  UnusedMethod() // не используется для SomeObject - не нужен
}

type SomeObject {
  i Interface
}

func (o * SomeObject) SomeMethod() {
  o.i.SomeMethod() // точно использованный метод интерфейса - stage1
}
```

---

Ах да, вы абсолютно правы! Я перевернул логику с ног на голову.

Главное преимущество анализа через AST в `linter.go` в том, что он быстро и надёжно находит ИСПОЛЬЗУЕМЫЕ методы в простых случаях:
1. Прямые вызовы методов через селектор: `obj.Method()`
2. Присваивания интерфейсов
3. Передача в функции как параметров
4. Возврат из функций

И только для тех методов, которые AST-анализ не смог точно определить как используемые, мы применяем подход из `try.go` с `staticcheck`. 

Это гораздо эффективнее, потому что:
1. Большинство методов используются простым и очевидным образом
2. Мы избегаем дорогой компиляции для явно используемых методов
3. `staticcheck` запускается только для "подозрительных" методов, где нужна полная проверка

То есть правильный алгоритм:
1. Сначала через AST находим все ТОЧНО используемые методы
2. Для оставшихся "подозрительных" методов используем подход с `staticcheck`

Спасибо за поправку! Это действительно принципиально меняет подход к оптимизации.

---

Давайте посмотрим на возможные упрощения в `linter.go`:

1. Избыточная обработка дженериков:
```go
// Сейчас:
if l.isGenericInterface(x) {
    l.addGenericWarning(x, interfaceType, filename, pkg.Fset)
    return true // Пропускаем анализ дженерик-интерфейсов
}
```
Мы можем вообще убрать эту логику, т.к. все сложные случаи (включая дженерики) будут обрабатываться через `staticcheck`

2. Сложная структура с предупреждениями:
```go
type GenericWarning struct {
    InterfaceName string
    File          string
    Line          int
    MethodCount   int
    TypeParams    string
}
```
Можно удалить, она нужна только для дженериков

3. Избыточные поля в `InterfaceMethod`:
```go
type InterfaceMethod struct {
    InterfaceName string
    MethodName    string
    Signature     string  // можно убрать, не используется для определения использования
    File          string
    Line          int
    Interface     *types.Interface // можно упростить, если не делать сложных сравнений
}
```

4. Сложная загрузка пакетов:
```go
cfg := &packages.Config{
    Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
        packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo |
        packages.NeedSyntax | packages.NeedModule,
    Dir: dir,
}
```
Можно оставить только то, что нужно для базового анализа: `NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax`

5. Сложная проверка использования методов:
```go
func (l *UnusedMethodLinter) isMethodUsed(method InterfaceMethod) bool {
    // Много разных проверок...
}
```
Можно оставить только самые базовые проверки:
- Прямые вызовы через селектор
- Присваивания интерфейсов
- Передача как параметров

6. Избыточные вспомогательные функции:
```go
func (l *UnusedMethodLinter) getTypeParamsString(...)
func (l *UnusedMethodLinter) getMethodSignature(...)
```
Можно удалить, они нужны только для вывода информации о дженериках

7. Verbose режим и логирование:
```go
if l.verbose {
    fmt.Printf("...")
}
```
Можно упростить или вообще убрать, т.к. основной вывод будет через `staticcheck`

8. Сравнение интерфейсов:
```go
func (l *UnusedMethodLinter) isSameInterface(iface1, iface2 *types.Interface) bool {
    // Сложная логика сравнения...
}
```
Можно упростить или удалить, если не делать сложных проверок совместимости

9. Конфигурация:
```go
type ConfigInterface interface {
    ShouldIgnore(filePath string) bool
}
```
Можно упростить до простого списка игнорируемых директорий, без интерфейса

В итоге, основная идея упрощения:
1. Оставить только базовый AST-анализ для очевидных случаев использования методов
2. Убрать всю логику обработки сложных случаев (они уйдут в `staticcheck`)
3. Упростить структуры данных и вспомогательные функции
4. Сфокусироваться на надежном определении ИСПОЛЬЗУЕМЫХ методов

Это сделает код проще, понятнее и легче в поддержке, при этом сохранив его основную полезную функцию.

---

О, отличная идея! Это максимально упростит код - мы ищем только прямые вызовы метода через селектор (`o.i.SomeMethod()`).

Такой подход:
1. Очень надёжный - если метод вызывается напрямую, он точно используется
2. Простой в реализации - нам нужно искать только `*ast.SelectorExpr`
3. Быстрый - не нужно анализировать типы, присваивания, и т.д.

Можно переписать `linter.go` примерно так:

```go
package linter

import (
    "go/ast"
    "go/parser"
    "go/token"
)

type Method struct {
    InterfaceName string
    MethodName    string
}

func FindUsedMethods(files []string) (map[Method]bool, error) {
    used := make(map[Method]bool)
    
    fset := token.NewFileSet()
    for _, file := range files {
        f, err := parser.ParseFile(fset, file, nil, 0)
        if err != nil {
            return nil, err
        }
        
        // Ищем прямые вызовы методов
        ast.Inspect(f, func(n ast.Node) bool {
            if sel, ok := n.(*ast.SelectorExpr); ok {
                if call, ok := sel.Parent.(*ast.CallExpr); ok {
                    // Нашли вызов метода
                    if x, ok := sel.X.(*ast.SelectorExpr); ok {
                        // Это вызов вида o.i.Method()
                        used[Method{
                            InterfaceName: x.Sel.Name,
                            MethodName: sel.Sel.Name,
                        }] = true
                    }
                }
            }
            return true
        })
    }
    
    return used, nil
}
```

А затем для оставшихся методов используем подход с `staticcheck`. Это даст нам:

1. Быстрое и надёжное определение используемых методов в простых случаях
2. Очень простой код без сложной логики
3. Гарантированно корректный результат через `staticcheck` для всех остальных случаев

Что думаете о таком радикальном упрощении?

----

Нет, не усложнит! В рамках одного пакета это всё тот же простой случай - прямой вызов метода. Просто расширим область поиска на все файлы пакета:

```go
package linter

import (
    "go/ast"
    "go/parser"
    "go/token"
    "path/filepath"
)

// Ищем o.i.Method() во всех файлах одного пакета
func FindUsedMethods(pkgDir string) (map[string]bool, error) {
    used := make(map[string]bool)
    fset := token.NewFileSet()
    
    // Читаем все .go файлы пакета
    files, err := filepath.Glob(filepath.Join(pkgDir, "*.go"))
    if err != nil {
        return nil, err
    }
    
    for _, file := range files {
        // Пропускаем тестовые файлы
        if filepath.Ext(file) == "_test.go" {
            continue
        }
        
        f, err := parser.ParseFile(fset, file, nil, 0)
        if err != nil {
            return nil, err
        }
        
        ast.Inspect(f, func(n ast.Node) bool {
            if call, ok := n.(*ast.CallExpr); ok {
                if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
                    used[sel.Sel.Name] = true
                }
            }
            return true
        })
    }
    
    return used, nil
}
```

Логика осталась такой же простой:
1. Тот же самый поиск прямых вызовов `o.i.Method()`
2. Просто проходим по всем файлам пакета
3. Всё остальное по-прежнему через `staticcheck`

Это хорошо вписывается в нашу стратегию:
- Быстро находим явно используемые методы в простых случаях
- Не усложняем код анализом типов и зависимостей
- Для всех сложных случаев используем `staticcheck`

---

`strings.HasSuffix(pkg.PkgPath, "_test")` - не нужен,у нас есть `config.ShouldIgnore()`

---

Да, можно оптимизировать поиск объявлений интерфейсов! Вместо обхода всего AST через `ast.Inspect`, мы можем сразу искать в `ast.File.Decls` объявления типов (`ast.GenDecl` с токеном `token.TYPE`), а в них - спецификации интерфейсов.

Давайте посмотрим текущий код:
Search files...
Да, действительно можно оптимизировать поиск объявлений интерфейсов. Вместо использования `ast.Inspect` для обхода всего AST, мы можем сразу искать в `ast.File.Decls` объявления типов. Вот как это можно сделать:
Преимущества этого подхода:

1. Более эффективный - мы не обходим весь AST, а сразу смотрим в нужные узлы:
   - `file.Decls` - все объявления в файле
   - Фильтруем только `*ast.GenDecl` с токеном `token.TYPE`
   - Проверяем только спецификации типов (`*ast.TypeSpec`)
   - Ищем только интерфейсы (`*ast.InterfaceType`)

2. Более понятный - код явно показывает, что мы ищем:
   - Объявления типов
   - Спецификации типов
   - Интерфейсы

3. Более структурированный - каждый шаг проверки выделен в отдельный блок:
   - Проверка типа объявления
   - Проверка спецификации
   - Проверка на интерфейс
   - Проверка на дженерики
   - Извлечение методов

4. Более поддерживаемый - легче добавлять новые проверки или модифицировать существующие, так как структура кода четко отражает структуру AST.

Это изменение не меняет функциональность, но делает код более эффективным и понятным.

----

Давайте разберем это подробнее:

1. `sel.Sel.Name` - это просто строка с именем метода или поля в выражении вида `x.y`. Например:
```go
foo.Bar()  // sel.Sel.Name == "Bar"
x.y.z()    // sel.Sel.Name == "z"
```

2. Только проверки `sel.Sel.Name` действительно недостаточно для методов с одинаковыми именами в разных интерфейсах. Давайте посмотрим пример:
Search files...
Из анализа кода видно, что только проверки `sel.Sel.Name` действительно недостаточно. Нам нужно учитывать:

1. Сигнатуру метода (параметры и возвращаемые значения)
2. Тип интерфейса, которому принадлежит метод

В коде есть несколько тестовых случаев, которые демонстрируют эту проблему:

1. Кейс 9 (ProcessorV1 и ProcessorV2):
```go
type ProcessorV1 interface {
    Process(data string) error // используется
}

type ProcessorV2 interface {
    Process(data string, options map[string]interface{}) error // не используется
}
```

2. Кейс 13 (Reader и AnotherReader):
```go
type Reader interface {
    CustomRead() error // используется через type assertion
}

type AnotherReader interface {
    CustomRead(data []byte) error // не используется, но имеет такое же имя
}
```

3. Кейс 14 (AnyProcessor и AnyHandler):
```go
type AnyProcessor interface {
    Process(data interface{}) error // используется
}

type AnyHandler interface {
    Process(data interface{}, options interface{}) error // не используется
}
```

Чтобы правильно определять использование методов, нам нужно:

1. Получить тип интерфейса через `types.TypeOf()`
2. Сравнивать не только имена методов, но и их сигнатуры
3. Учитывать контекст вызова (через какой интерфейс вызывается метод)

Это позволит различать методы с одинаковыми именами, но разными сигнатурами в разных интерфейсах.

---

Давайте проанализирую, насколько сильно усложнится код:

1. Нам нужно будет добавить информацию о типах:
```go
type InterfaceMethod struct {
    InterfaceName string
    MethodName    string
    Interface     *types.Interface  // уже есть
    Method        *types.Func       // нужно добавить
}
```

2. При поиске вызовов нужно будет:
```go
if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
    // Текущий код - только проверка имени
    methodName := sel.Sel.Name
    
    // Нужно добавить:
    // 1. Получить тип receiver'а
    receiverType := pkg.TypesInfo.TypeOf(sel.X)
    // 2. Получить тип метода
    methodType := pkg.TypesInfo.TypeOf(sel.Sel)
    // 3. Сравнить сигнатуры
    if isSameSignature(methodType, interfaceMethod.Method) {
        // метод используется
    }
}
```

На самом деле, это не сильное усложнение, потому что:

1. Большая часть инфраструктуры уже есть - мы уже работаем с `types.Package` и `types.Interface`
2. Go предоставляет удобные методы для работы с типами через `go/types`
3. Сравнение сигнатур - это стандартная операция в Go

Плюсы от добавления этой проверки:
1. Более точное определение использования методов
2. Меньше ложных срабатываний
3. Правильная работа с одноименными методами

Я бы сказал, что усложнение оправдано, учитывая повышение точности анализа.

---

Да, абсолютно верно! Давайте рассмотрим пример:

```go
// Два разных интерфейса с одинаковыми методами
type Storage interface {
    Get(key string) (string, error)
}

type Cache interface {
    Get(key string) (string, error)  // Абсолютно такая же сигнатура!
}

func main() {
    var s Storage = &MyStorage{}
    s.Get("key")  // Используется метод из Storage

    // Cache.Get может быть неиспользуемым, хотя имя и сигнатура те же!
}
```

Значит для точного определения используемого метода нам нужно:
1. Имя метода
2. Сигнатура (параметры и возвращаемые значения)
3. Тип интерфейса, которому принадлежит метод

То есть когда мы находим вызов:
```go
if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
    // 1. Имя метода
    methodName := sel.Sel.Name
    
    // 2. Тип receiver'а - нужен чтобы понять, какой интерфейс
    receiverType := pkg.TypesInfo.TypeOf(sel.X)
    
    // 3. Сигнатура метода
    methodType := pkg.TypesInfo.TypeOf(sel.Sel)
}
```

Только если все три параметра совпадают, мы можем быть уверены, что нашли именно тот метод, который используется. Иначе можем пропустить неиспользуемые методы в других интерфейсах.

Получается, что проверка типов все-таки нужна для точного определения используемых методов.

----

на stage1 нужно составлять мапу подозрительный методов **по пакетам** для stage2

--- 

`filepath.Join()` - обязательно использовать для кросплатформенности

---

упрощать конфигурацию не надо, используем готовый `pkg/config`, только надо применить `filepath.Join()`

---

чтобы не спотыкаться на `USED:` VS `UNUSED:` (которая частично содержит `USED:`), переименуем USED на `OK:`

--- 

как мы это будем делать по структуре проекта? оставим весь старый код как есть (будем туда подсматривать за поиском вдохновения и готовых решений). 

### v1
cmd/v1/main.go
pkg/linter/linter.go
pkg/linter/linter.go
pkg/config/config.go

### v2
cmd/v2/main.go
pkg/stage1/stage1.go
pkg/stage2/stage2.go
pkg/config/config.go - будет ломающих использовать без переделок

### доделаем тесты для применения в v2 (на v1 они всё равно не работали)
test/unused/unused_test.go
test/unused/unused.1_test.go
test/unused/unused.2_test.go
test/data - все тесты (и юнит и приёмочные) должны использовать тестовые данные, а не создавать для себя синтетику

---

- stage1 - упрощенный обработчик AST для проверки базового кейса - точно использованных методов интерфейсов, и собирает их в мапу по пакетам для исключения повторной обработки на stage2

- stage2 - то, что мы сделали в pkg/try, проверяет методы через удаление в файле с проверкой через staticcheck, и сообщает о неиспользованных методах 

---

```go
				// TODO: можно не восстанавливать начальное состоятие,
				// пока удаление метода не ломает staticcheck
				err = writeStringToFile(filepath.Join(tmpDir, file), content)
				if err != nil {
					return nil, fmt.Errorf("failed to restore file: %v", err)
				}
```

---

UNIX-WAY!!

---

мы уже всё сделали. на этапе тестирования приёмочных тестов, которые в - test/unused - должны использовать данные из test/data. запускай тесты и добивайся, чтобы они все были правильно реализованы и работали. нумерация тестов и их структура соответсвует test/data/interfaces.go

---

Отлично! Тест проходит, потому что теперь он правильно отражает логику stage1:
Stage1 находит все прямые вызовы методов через селектор o.i.Method()
Не важно где происходит вызов (в type switch, горутине, defer) - если это синтаксически o.i.Method(), то stage1 должен его найти
Это делает stage1 очень простым и надежным - он просто ищет синтаксическую конструкцию, не заботясь о контексте вызова.
