DelayedNotifier

🚀 Установка и запуск
 
 Запустите проект:

    1) make dockerRun - ВСЕ УСТАНОВИТСЯ

🧪 Тестирование

Чтобы затестировать код и узнать покрытие введите команду:

    make testAll

💻 Технический стек

    Язык программирования: Go 1.25
    Web-фреймворк: Ginext
    База данных: PostgreSQL
    Миграции: Migrate
    Логгирование: zerolog
    Документация API: Swagger
    Кэш: Redis 
    Контейнеризация: Docker, Docker Compose

HTTP API

notifyGroup := r.Group("/notify")
	{
		notifyGroup.POST("/create", h.CreateHanlder)
		notifyGroup.GET("/all", h.GetAllHandler)
		notifyGroup.GET("/status/:id", h.StatusHanlder)
		notifyGroup.DELETE("/cancel/:id", h.CancelHanlder)
	}







