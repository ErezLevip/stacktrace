# stacktrace

Stacktrace is a simple library that  gives you the ability to quickly log stacktrace.

## How-to

as a string:

	trace := stacktrace.NewStackTrace(0)
	tj, err := trace.ToJson()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(tj, err)
 
as a reader:

	trace := stacktrace.NewStackTrace(0)
	b, err := ioutil.ReadAll(trace)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(string(b), err)
 
