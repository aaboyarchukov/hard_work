# Антипаттерн "Самодокументирующийся код"

## Введение

После прочтения материала - Антипаттерн "Самодокументирующийся код", я для себя понял что, безусловно, необходимо писать читаемый, чистый и ясный код, но зачастую, даже при самой чистой выразительности кода, разработчикам, которые не погружены в доменную область определенного проекта, достаточно сложно сразу разобраться с кодовой базой. А хороший, емкий и выразительный комментарий, может ускорить процесс погружения в проект! То есть оставлять комментарии - является признаком хорошего тона и понимания по отношению к своим коллегам и не только. Но важно помнить, какие именно оставлять комментарии и с какой частотой!

Комментарии не должны покрывать большее количество вашего кода, их необходимо писать именно в тех местах, которые исполняют главную роль в вашем проекте, а также много где используются. Также при написании комментариев, важно думать именно на 3 логическом уровне рассуждений и программе, то есть комментарии должны описывать спецификацию вашей программы, а не ее исполнение/компиляцию.
## Пример 1

В данном примере я комментирую важную часть проекта, которая отвечает за асинхронную работу со страховой компанией, выстраивая взаимосвязь с ней. Думаю, что разработчику важно об этом знать при погружении в проект и кодовую базу. 

```go
// Компонент - Worker организовывает асинхронную работу со страховой
type Worker struct {
	identification *IdentificationWorker
	outbox         *OutboxWorker
}

func New(identification *IdentificationWorker, outbox *OutboxWorker) *Worker {
	return &Worker{
		identification: identification,
		outbox:         outbox,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return w.identification.Run(ctx)
	})

	g.Go(func() error {
		return w.outbox.Run(ctx)
	})

	return g.Wait()
}

```


## Пример 2

В данном примере был объяснен архитектурный подход для построения приложения, чтобы разработчики писали единообразно. Это могло быть в документации, но люди чаще читают код, чем документацию :)

```go
// Container содержит все сервисы приложения
// используется композиция для 
// разделения имплементации отдельных сервисов по пакетам
type Container struct {
	Sessions      Sessions
	Auth          Auth
	Users         Users
	Notifications Notifications
	Payments      Payments
}

func NewContainer(sessions Sessions, auth Auth, users Users, notifications Notifications, payments Payments) *Container {
	return &Container{
		Sessions:      sessions,
		Auth:          auth,
		Users:         users,
		Notifications: notifications,
		Payments:      payments,
	}
}


```

## Пример 3

В данном примере комментарий необходим, чтобы донести разработчику смысл использования SSE в текущем проекте.

```go
// Компонент - FlushingSSEResponse, отправляет подготовленные 
// буферризированные сообщения
// подключенным пользователям и очищает буфер.
// Важно! При работе с SSE всегда важно использовать Flush:
// https://developer.mozilla.org/ru/docs/Web/API/Server-sent_events/Using_server-sent_events
type FlushingSSEResponse struct {
	stream *SSEStream
}

// Компонент - SSEStream, поток для упраления передачей сообщений по каналу.
// Используется pipereader для синхронной передачи информации 
// между писателем и читателем
type SSEStream struct {
	reader *io.PipeReader
}

// Компонент - SseUser, пользователь, который подключен к каналу по SSE
type SseUser struct {
	UserID  domain.UserID
	MsgChan chan dto.Notification
}

// Компонент - SseHub, хранилище, которое хранит, подключенный пользователей
type SseHub struct {
	usersHub map[domain.UserID]*SseUser
	mx       *sync.RWMutex
}

// Копмонент - Worker, отвечает за отправку и обработку увеодмлений
// для подключенных пользователей
type Worker struct {
	service Service
	timer   *time.Ticker
}
func NewNotificationsWorker(notificationsService Service) *Worker {
	newTimer := time.NewTicker(time.Duration(time.Second * 1))
	return &Worker{
		timer:   newTimer,
		service: notificationsService,
	}
}

```