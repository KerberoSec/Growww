package src

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	validEnvironments = map[string]bool{"prod": true, "staging": true, "testnet": true, "dev": true}
	validEntities     = map[string]bool{"domestic": true, "giftcity": true, "global": true}
	topicRegex        = regexp.MustCompile(`^[a-z0-9_]+$`)
	versionRegex      = regexp.MustCompile(`^v[1-9][0-9]*$`)
)

// ParsedTopic represents validated Kafka topic segments
type ParsedTopic struct {
	Environment    string
	Entity         string
	BoundedContext string
	Aggregate      string
	EventName      string
	Version        string
}

// ValidateTopicName validates topic adheres to <env>.<entity>.<bounded-context>.<aggregate>.<event-name>.<version>
func ValidateTopicName(topic string) (*ParsedTopic, error) {
	if topic == "" {
		return nil, errors.New("topic cannot be empty")
	}

	parts := strings.Split(topic, ".")
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid topic grammar: expected 6 segments, got %d in '%s'", len(parts), topic)
	}

	env, entity, context, aggregate, event, version := parts[0], parts[1], parts[2], parts[3], parts[4], parts[5]

	if !validEnvironments[env] {
		return nil, fmt.Errorf("invalid environment '%s'", env)
	}
	if !validEntities[entity] {
		return nil, fmt.Errorf("invalid entity '%s'", entity)
	}
	if !topicRegex.MatchString(context) {
		return nil, fmt.Errorf("invalid bounded context '%s'", context)
	}
	if !topicRegex.MatchString(aggregate) {
		return nil, fmt.Errorf("invalid aggregate '%s'", aggregate)
	}
	if !topicRegex.MatchString(event) {
		return nil, fmt.Errorf("invalid event name '%s'", event)
	}
	if !versionRegex.MatchString(version) {
		return nil, fmt.Errorf("invalid version '%s' (expected v1, v2, etc.)", version)
	}

	return &ParsedTopic{
		Environment:    env,
		Entity:         entity,
		BoundedContext: context,
		Aggregate:      aggregate,
		EventName:      event,
		Version:        version,
	}, nil
}

// SelectPartitionKey determines the deterministic partition key based on domain context
func SelectPartitionKey(context string, entityID, symbol, settlementID string) (string, error) {
	switch strings.ToLower(context) {
	case "trading", "orders", "balances", "kyc":
		if entityID == "" {
			return "", errors.New("user/account entityID required for user partition")
		}
		return entityID, nil
	case "matching", "marketdata", "quotes":
		if symbol == "" {
			return "", errors.New("symbol required for market partition")
		}
		return symbol, nil
	case "settlement", "dvp", "clearing":
		if settlementID == "" {
			return "", errors.New("settlementID required for clearing partition")
		}
		return settlementID, nil
	default:
		if entityID != "" {
			return entityID, nil
		}
		return "default_partition", nil
	}
}
