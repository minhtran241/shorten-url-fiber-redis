package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	googleuuid "github.com/google/uuid"

	"github.com/minhtran241/shorten-url-fiber-redis/database"
	"github.com/minhtran241/shorten-url-fiber-redis/helpers"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"custom_short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"custom_short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON format",
		})
	}

	// implement rate limiting

	// open db for rate limiting of the IP address
	r2 := database.CreateClient(1)
	defer r2.Close()
	// get the remaining access of the IP address
	rateRemaining, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil { // not found -> new IP, user -> provide new API_QUOTA and 30 minutes expiry
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*time.Minute).Err()
	} else { // found
		valInt, _ := strconv.Atoi(rateRemaining)
		if valInt <= 0 { // if reach rate limit restriction
			// get the remaining time to reset (TTL of the pair)
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":           "rate limit exceeded",
				"rate_limit_rest": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	// check if the input is an actual URL
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid URL",
		})
	}

	// check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "domain error: infinity loop detected",
		})
	}

	// enforce https, SSL
	body.URL = helpers.EnforceHTTP(body.URL)

	// the key (custom URL or some random uuid)
	var id string

	if body.CustomShort == "" { // if no custom short provided
		// set id to random uuid
		id = googleuuid.New().String()[:6]
	} else { // custom short provided
		// set id to custom short provided
		id = body.CustomShort
	}

	// open db for shortened URL
	r := database.CreateClient(0)
	defer r.Close()

	actualURL, _ := r.Get(database.Ctx, id).Result()
	if actualURL != "" { // if the custom short is already used
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "URL custom short is already used",
		})
	}

	// set the expiry of this custom short to 24 hours
	if body.Expiry == 0 {
		body.Expiry = 24
	}

	// set key-value pair custom_short & actual URL
	err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()

	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unable to connect to server",
		})
	}

	res := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	// decrement the remaining access of the IP
	r2.Decr(database.Ctx, c.IP())

	// the current remaining access of the IP
	rateRemaining, _ = r2.Get(database.Ctx, c.IP()).Result()
	// set XRateLimitRemaining
	res.XRateRemaining, _ = strconv.Atoi(rateRemaining)
	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	// set the remaining time to reset by getting the TTL of the IP
	res.XRateLimitReset = ttl / time.Nanosecond / time.Minute
	// custom_short = domain/id
	res.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(res)
}
