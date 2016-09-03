package route

func Init() {
	initReverseProxy()
	initStatic()
	initAPI()
}
