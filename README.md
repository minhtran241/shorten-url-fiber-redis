## URL Shortener Service

![Fiber](https://img.shields.io/badge/Go%20-%20Fiber-00ADD8?style=flat-square&logo=go&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-DC382D?style=flat-square&logo=redis&logoColor=white)

### Data format

-   Request format

    ```
    type request struct {
        URL         string        `json:"url"`
        CustomShort string        `json:"custom_short"`
        Expiry      time.Duration `json:"expiry"`
    }
    ```

-   Response format
    ```
    type response struct {
        URL             string        `json:"url"`
        CustomShort     string        `json:"custom_short"`
        Expiry          time.Duration `json:"expiry"`
        XRateRemaining  int           `json:"rate_limit"`
        XRateLimitReset time.Duration `json:"rate_limit_reset"`
    }
    ```

### Redis Server

```
rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("DATABASE_ADDRESS"),
		Password: os.Getenv("DATABASE_PASSWORD"),
		DB:       dbNo,
	})
```

The Service uses 2 Redis as databases

-   The first database

    -   Restricts the number of access to the service by one IP address
    -   Stores the remaining number of access of IP addresses that send requests to the server
    -   The default number of access to the service is 10 in 30 minutes

-   The second database
    -   Store shortened URLs with actual URLs
    -   If shortened URL is not provided in the request, a random `uuid` will be generated to become the `id (key)` for the actual URL
    -   If the shortened URL in the request is already used and is not expired in Redis, user can not use this shortened version
    -   The default duration for one shortened URL to exist in Redis is 24 hours

### Package index

List dependencies used in this system by using [github.com/ribice/glice](https://github.com/ribice/glice), a Golang license and dependency checker. The package prints list of all dependencies, their URL, license and saves all the license files in /licenses

```
+-----------------------------------+-------------------------------------------+--------------+
|            DEPENDENCY             |                  REPOURL                  |   LICENSE    |
+-----------------------------------+-------------------------------------------+--------------+
| github.com/asaskevich/govalidator | https://github.com/asaskevich/govalidator | MIT          |
| github.com/go-redis/redis/v8      | https://github.com/go-redis/redis         | bsd-2-clause |
| github.com/gofiber/fiber/v2       | https://github.com/gofiber/fiber          | MIT          |
| github.com/google/uuid            | https://github.com/google/uuid            | bsd-3-clause |
| github.com/joho/godotenv          | https://github.com/joho/godotenv          | MIT          |
+-----------------------------------+-------------------------------------------+--------------+
```

### Endpoints

-   Shorten URL
    -   Endpoint: `/api/v1`
    -   Method: `POST`
    -   Input: Client passes in the new object contains information of the shortened URL via JSON body (go to the `Sample request and response` section to see sample request)
    -   Output: A new object is created in the database
-   Resolve URL
    -   Endpoint: `/:url`
    -   Method: `GET`
    -   Input: Client passes in the shortened version of the actual URL they created
    -   Output: Client will be redirected to the actual URL if the shortened version exists in the database

### Sample request and response

-   Request

```
{
    "url": "https://redis.io/commands/ttl/",
    "custom_short": "ttl",
    "expiry": 50
}
```

-   Response:

```
{
    "url": "https://redis.io/commands/ttl/",
    "custom_short": "localhost:3000/ttl",
    "expiry": 50,
    "rate_limit": 9,
    "rate_limit_reset": 30
}
```
