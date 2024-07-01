<br />
<p align="center">

  <img src="https://github.com/devzery/devzery-go-gorilla-mux/assets/86240862/5081cc82-7389-401e-99f3-fd440ce630ad" alt="Logo" width="300" height="">



  <h3 align="center">Devzery Go SDK</h3>

  <p align="center">
    Test your API with AI
    <br />

</p>

# Devzery Go SDK

Devzery's Go SDK helps you test your API using the power of AI. 
Use Devzery to achieve end to end testing of your API by just adding few lines of code

## Installation

Install the package by running the following command in your terminal

```bash
  go get github.com/devzery/devzery-go-gorilla-mux
```



## Quick Start

Paste the following code in your router

```go
	func main(){
	    r := mux.NewRouter() //Here replace 'r' with your router name
		mw := loggingMiddleware.create(
			"API_ENDPOINT",
			"YOUR_ACCESS_TOKEN",
			"YOUR_SOURCE_NAME",
		)
		r.Use(mw.LoggingMiddleware)
	}
```
"YOUR_SOURCE_NAME" should be the name of the service which is hitting our API_ENDPOINT
<br/>
<br/>
You can now send requests to your APIs and Devzery will take care of testing on them. Yes it is that simple!
