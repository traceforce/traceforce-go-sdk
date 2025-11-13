# Traceforce Go SDK

A Go SDK for interacting with the Traceforce API to manage connections and other resources.

## Installation
First create an API client on the Traceforce UI. Then get an API key using the following API
```
POST https://www.traceforce.co/api/v1/api-keys
{
    "client_id": <client_id_from_the_ui>,
    "client_secret": "<client_secret_from_the_ui>"
}
```


## Usage

### Initialize the Client
```
client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
```

