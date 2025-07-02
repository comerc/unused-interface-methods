package test_data

// ===============================
// ДЖЕНЕРИК-ИНТЕРФЕЙСЫ
// ===============================

// 1. Простой дженерик
type SimpleRepo[T any] interface {
	Get(id string) T
	Save(item T) error
}

// 2. Дженерик с ограничениями
type Comparable interface {
	Compare(other Comparable) int
}

type SortableRepo[T Comparable] interface {
	GetSorted() []T
	Insert(item T) error
	Remove(item T) bool
}

// 3. Множественные параметры типа
type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	Delete(key K) bool
	Keys() []K
	Values() []V
}

// 4. Сложные ограничения
type Serializable interface {
	Serialize() []byte
	Deserialize([]byte) error
}

type PersistentCache[K comparable, V Serializable] interface {
	Load(key K) (V, error)
	Store(key K, value V) error
	Persist() error
	Restore() error
}

// 5. Вложенные дженерики
type NestedRepo[T any] interface {
	GetMap() map[string]T
	GetSlice() []T
	GetChannel() chan T
	ProcessBatch(items []T) (map[string]T, error)
}

// 6. Дженерик-интерфейс из doc/GENERICS_PROBLEM.md
type GenericRepository[T any] interface {
	Get(id string) (T, error) // используется
	Save(item T) error        // используется
	Delete(id string) error   // не используется
	List() ([]T, error)       // используется
}

// 7. Простой дженерик для тестирования
type Repository[T any] interface {
	Get(id string) (T, error) // используется
	Save(item T) error        // не используется
	Delete(id string) error   // не используется
	List() ([]T, error)       // используется
}

// ===============================
// ОБЫЧНЫЙ ИНТЕРФЕЙС ДЛЯ СРАВНЕНИЯ
// ===============================

// Обычный интерфейс для сравнения с дженериками
type RegularInterface interface {
	DoSomething() error // используется
	GetResult() string  // не используется
}

// ===============================
// ИСПОЛЬЗОВАНИЕ
// ===============================

// Конкретные типы
type User struct {
	ID   string
	Name string
}

type Post struct {
	ID    string
	Title string
}

// Использование дженериков
type UserService struct {
	userRepo GenericRepository[User]
	repo     Repository[User]
}

type PostService struct {
	postRepo GenericRepository[Post]
	repo     Repository[Post]
}

// Использование обычного интерфейса
type Service struct {
	regular RegularInterface
}

func (s *Service) Work() {
	s.regular.DoSomething() // Этот метод используется
	// GetResult() НЕ используется
}

// Использование дженерик-методов
func (us *UserService) GetUser(id string) (*User, error) {
	// Вызов на GenericRepository[User]
	user, err := us.userRepo.Get(id)
	if err != nil {
		return nil, err
	}

	// Вызов на Repository[User]
	us.repo.Get(id)

	return &user, nil
}

func (us *UserService) SaveUser(user User) error {
	// Вызов Save на GenericRepository[User]
	return us.userRepo.Save(user)
}

func (ps *PostService) ListPosts() ([]Post, error) {
	// Вызов List на GenericRepository[Post]
	posts, err := ps.postRepo.List()
	if err != nil {
		return nil, err
	}

	// Вызов на Repository[Post]
	ps.repo.List()

	return posts, nil
}

// Delete НЕ используется ни в одном инстанцировании
// Save в Repository[T] НЕ используется
