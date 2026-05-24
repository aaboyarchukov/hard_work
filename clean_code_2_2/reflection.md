# Ясный код 2-2

Рассмотрим примеры плохого проектирования классов на уровне классов и на уровне приложения:

## Уровень классов

1. Класс слишком большой (нарушение SRP), или в программе создаётся слишком много его инстансов (подумайте, почему это плохой признак).

Такое может возникать при плохом проектировании и отсутствии знаний базовых принципов при построении программы (чистая и ясная архитектура).
```go
type Repository interface {
	// CRUD for user
	GetUser(ctx context.Context, id int) (model.User, error)
	UpdateUser(ctx context.Context, opts ...opts.User) error
	DeleteUser(ctx context.Context, id int) error
	SaveUser(ctx context.Context, user model.User) error
	
	// CRUD for task
	GetTask(ctx context.Context, id int) (model.Task, error)
	UpdateTask(ctx context.Context, opts ...opts.Task) error
	DeleteTask(ctx context.Context, id int) error
	SaveTask(ctx context.Context, user model.User) error
	
	// ...
}
```
Интерфейс репозитория захламлен, он ответственен как за пользователей, так и за задачи, а также за много другое. Такое необходимо разделять на четкие, ясные и атомарные сущности, которые ответственны за одну вещь (работа только с задачами, работа только с пользователями и тд)

2. Класс слишком маленький или делает слишком мало.
Такое может возникнуть при чресчур сильным разделением, что в итоге приводит к абсолютно бесмысленному разделению, когда класс делает слишком мало.
```go
// delete_user_repository.go
type DeleteUserRepository interface {
	DeleteUser(ctx context.Context, id int) error
}


// save_user_repository.go
type SaveUserRepository interface {
	SaveUser(ctx context.Context, user model.User) error
}


// get_user_repository.go
type GetUserRepository interface {
	GetUser(ctx context.Context, id int) (model.User, error)
}
```
Такие классы можно объединить в один, которй будет полностью ответственен за управление пользователями.
3. В классе есть метод, который выглядит более подходящим для другого класса. 
Опять же это проявляется при нарушении SRP и плохом проектировании
```python
class Bike:
	def __init__(self):
		pass
	
	def ride(): ...
	def stop(): ...
	def get_speed(): ...
	
	# ???
	def fix(): ...
	
```
Логически можно понять, что метод починки велосипеда никак не может относится к нему напрямую. Для этого должен быть класс Ремонтник, с возможностью починки любого транспорта.

4. Класс хранит данные, которые загоняются в него в множестве разных мест в программе.
В данном пункте увеличивается связность между классами, что неизбежно ведет к сложному рефакторингу, масштабированию и изменениям в целом.
```python
class OrderService:
	def __init__(self, fav_orders, certain_orders, orders_queue):
		self.fav_orders = fav_orders # from repo
		self.certain_orders = certain_orders # from repo
		self.orders_queue = orders_queue # from queue manager
```
Класс сильно связывается с другими зависимостями, что увеличивает coupling. По хорошему классы должны являться провайдерами данных, никак не прямым хранилищем.

5. Класс зависит от деталей реализации других классов.
Здесь также возникают случаи из-за увеличения связности между классами, что помимо предыдущих пунктов - ведет к сложности в тестировании, ведь как возможно проверить атомарную единицу в виде конкретного класса, если он зависит от другого - они проверяются в цепочке.
В данном примере мы слишком сильно разбили составляющие класса `Bike`, таким образом при малейшем изменении зависимостей - необходимо менять и сам класс `Bike`
```python
class Bike:
	def __init__(self, engine: Engine, body: BikeBody, pedals: Pedals):
		self.engine = engine
		self.body = body
		self.pedals = pedals
```

6. Приведение типов вниз по иерархии (родительские классы приводятся к дочерним).
В данном случае подойдет, конечно, если мы передадим родительский тип `Animal`, но то что мы конкретно передаем в аргументы `Cat`: делает возможным допустить ошибку при вызове методов, так как контракты у них будут различны
```python
def tame(self, animal: Cat, food: Food):
	self.feed(animal, food)
	self.stroke(animal)
	self.play_with_animal(animal)
	
	# ...some logic
	animal.drink_milk() # error!!!
```

7. Когда создаётся класс-наследник для какого-то класса, приходится создавать классы-наследники и для некоторых других классов.
Такое может возникнуть из-за сильной связности между классами, так, что изменение/внедрение, например, дочернего класса, ведет к изменению других классов, что не очень хорошо.
```python
class Engine:
	def __init__(self):
		pass

class ElectricEngine(Egine):
	def __init__(self):
		pass
		
class Bike:
	def __init__(self, engine: Engine):
		pass
		
class ElectricBike(Bike):
	def __init__(self, engine: ElectricEngine):
		pass
```
Тут мы в зависимости от типа байка, также добавили тип двигателя, в нашей нинешней архитектуре, также надо будет при каждом новом подклассе Байка, надо создавать еще подкласс к Двигателю 
8. Дочерние классы не используют методы и атрибуты родительских классов, или переопределяют родительские методы.

Если дочерние классы не переиспользуют методы и атрибуты родительских - явный признак, указывающий на плохое проектирование. Так как наша цель - добиться максимального повторного использования.
```python
class Duck:
	def __init__(self):
		pass
	
	def sound():
		print("кря")

class RubberDuck(Duck):
	def __init__(self):
		pass
	def sound(): ... # not implement

class Drake(Duck):
	def __init__(self):
		pass
	def sound():
		print("кря, я селезень")
```

## Уровень приложения

1. Одна модификация требует внесения изменений в несколько классов.
Такое происходит, при сильной связи в иерархии между модулями, например, модуль `Watcher` реализует паттерн Наблюдатель, но реализовали его неверно и криво - в этом примере модуль просто передается как зависимость, завязывая на себе другие модули, которые используют его, теперь при малейшем изменении класса `Watcher` - как по струнам, будут реагировать другие модули и требовать также изменений
```python
# whatcher.py
class Watcher:
	def __init__(self, chan_msgs):
		self.chan_msgs = chan_msgs
	
	def notify(self, event):
		self.chan_msgs.append(event)
		# notify
		

# email.py
import watcher
import smtp

class EmailSender:
	def __init__(self):
		pass
	
	def send(self, msg, dest_email):
		smtp.send(msg, dest_email)		
		watcher.notify(Event(...))
		
# sms.py
import watcher
import smpp

class SmsSender:
	def __init__(self):
		pass
	
	def send(self, msg, dest_phone):
		smpp.send(msg, dest_phone)		
		watcher.notify(Event(...))
```

2. Использование сложных паттернов проектирования там, где можно использовать более простой и незамысловатый дизайн.
Такие случи возникают при излишней рефлексии. Если например, у нас система не будет меняться сильно в будущем, то не имеет смысла прибегать к сильным паттернам для реализации какой-то логики, достаточно применить что-то лаконичное и простое. Например вводят паттерн `Strategy`, когда приложение ограничено лишь одной вариацией, тем самым не придерживаясь принципа `YAGNI`, просто оверинжениринг на ровном месте. Хотя мы знаем, что наш сервис не будет меняться и достаточно одного метода оплаты `CreditCard`
```go
type Payment interface {
	pay(ctx context.Context, amount float64, dest string) error
}

type PayPalService struct {
}

func (pps *PayPalService) pay(ctx context.Context, amount float64, dest string) error {
	// ...
}

type AliPayService struct {
}

func (aps *AliPayService) pay(ctx context.Context, amount float64, dest string) error {
	// ...
}

type YooMoneyService struct {
}

func (yms *YooMoneyService) pay(ctx context.Context, amount float64, dest string) error {
	// ...
}
```