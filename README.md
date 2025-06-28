# Traceforce Go SDK

A Go SDK for interacting with the Traceforce API to manage connections and other resources.

## Installation
Obtain an API token from the Traceforce UI or using the following API
```
POST https://www.traceforce.co/api/v1/api-keys
```


## Usage

### Initialize the Client
```
client, err := NewClient(os.Getenv("TRACEFORCE_API_KEY"), "", nil)
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
```

