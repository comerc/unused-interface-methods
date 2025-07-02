# Проблема анализа дженерик-интерфейсов

> Этот документ содержит детальное объяснение того, почему наш линтер не может корректно анализировать дженерик-интерфейсы.
> Примеры дженерик-интерфейсов и их использование находятся в `generics.go`

## Что видит линтер

1. Интерфейс `GenericRepository[T]` с методами:
   - `Get(id string) (T, error)`
   - `Save(item T) error`
   - `Delete(id string) error`
   - `List() ([]T, error)`

2. Переменные типов:
   - `us.userRepo` имеет тип `GenericRepository[User]`
   - `ps.postRepo` имеет тип `GenericRepository[Post]`

3. Вызовы методов:
   ```go
   us.userRepo.Get(id)     // на типе GenericRepository[User]
   us.userRepo.Save(user)  // на типе GenericRepository[User]
   ps.postRepo.List()      // на типе GenericRepository[Post]
   ```

## Что происходит в types.TypeOf()

```go
// types.TypeOf(us.userRepo) возвращает:
*types.Named {
  obj: *types.TypeName "GenericRepository"
  targs: []*types.Named{User}  // аргументы типа
  underlying: *types.Interface // базовый интерфейс
}

// types.TypeOf(ps.postRepo) возвращает:
*types.Named {
  obj: *types.TypeName "GenericRepository"
  targs: []*types.Named{Post}  // ДРУГИЕ аргументы типа
  underlying: *types.Interface // тот же базовый интерфейс
}
```

## Проблема в isMethodCallOnInterface()

```go
// Наш метод делает:
1. exprType := pkg.TypesInfo.TypeOf(sel.X)
   // exprType = GenericRepository[User] или GenericRepository[Post]

2. types.Implements(exprType, method.Interface)
   // method.Interface = базовый GenericRepository[T]
   // exprType = GenericRepository[User]
   // Результат: false! Потому что GenericRepository[User] ≠ GenericRepository[T]
```

## Решение

Нужно сравнивать не конкретные инстанцирования, а базовые типы.

Пример решения:
```go
func isGenericInterface(t types.Type) (*types.Named, bool) { ... }
func compareGenericInterfaces(t1, t2 types.Type) bool { ... }
```

## Дополнительные сложности

1. Вложенные дженерики:
   - `GenericRepository[map[string]User]`
   - `GenericRepository[[]Post]`

2. Ограничения типов:
   ```go
   type GenericRepository[T Serializable] interface { ... }
   ```

3. Множественные параметры типа:
   ```go
   type Cache[K comparable, V any] interface { ... }
   ```

4. Методы с дженерик-параметрами:
   ```go
   type Converter interface {
       Convert[T any](input string) (T, error)
   }
   ```

## Почему это сложно

1. **Система типов Go** - дженерики добавили новый уровень сложности
2. **Инстанцирование** - один интерфейс → множество конкретных типов
3. **Совместимость типов** - нужно понимать отношения между базовым типом и инстанцированиями
4. **API golang.org/x/tools** - требует глубокого понимания работы с дженериками

### Для полной поддержки дженериков нужно:
- Определять, что тип является дженериком
- Извлекать базовый тип без аргументов
- Сравнивать базовые типы, а не конкретные инстанцирования
- Обрабатывать сложные случаи с вложенными дженериками 