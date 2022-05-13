# cfServiceAlert
cfServiceAlert generates alerts for cloudfoundry service instances. Cloudfoundry platform users will receive alerts for services running in their org. This tool works in conjunction with cfNotificationService. Alerts are configured through rules.json (see below). Metrics are retrieved from prometheus. So you need to have prometheus in place for this tool to work.

# Running cfServiceAlert
This tool is developed to run on cloudfoundry. It won't run anywhere else. You can use the provided manifest.yml to deploy. Be sure to set the following env vars:

- CF_USER (user with readonly-admin permissions on CF)
- CF_PASSWORD (password for read-only admin user)
- CF_CLIENT (UAA client with readonly-admin permissions. This is an alternative for user. If CF_CLIENT is set this will be used instead of the user)
- CF_SECRET (Secret for the UAA Client)
- RULES_PATH (optional, path to the rules file. default: rules.json)
- PROMETHEUS_URL (required, URL for prometheus server, only tested with grafana datasrouce proxy url. Url look like: https://<grafana url/api/datasources/proxy/<id/)
- GRAFANA_API_KEY (required, API key for grafana. We assume we access prometheus through te grafana datasource proxy)
- CHECK_INTERVAL (optional, how often do we retrieve the metrics from prometheus in second. default: 120 seconds)
- ENVIRONMENT_NAME (required, Tells notificationService from which env this message is originating. Must match one of the configured env names in cfNotificationService)
- NOTIFICATION_URL (required, URL of the cfNotificationService)
- NOTIFICATION_USER (required, API user for cfNotificaitonService)
- NOTIFICATION_PASSWORD (required, Password for the API user)

# rules.json
Alerts are configured in rules.json. This repo contains an example rules.json. Here is some explenation:

```
{
    "<service name as known in CF>": [ {
        "name": "<alert name>",
        "prometheus_query": "<prometheus query>",
        "treshold": "<alert treshold>",
        "notification_interval": "<how often do we repeat the alert if the problem persists>",
        "above": <true will trigger alert if value is above treshold. False will trigger alert when value is below treshold>,
        "subject": "<golang template for the subject of the alert message> Check "GenerateMEssageForSpace" method in "alertRule.go" for variables that get exposed to the template>",
        "message": "<golang template for the body of the alert message> Check "GenerateMEssageForSpace" method in "alertRule.go" for variables that get exposed to the template>."
    },
    {
        "name": "<another alert name>",
        ....
        <you can have multiple alerts per service. For example: one for disk space, one for RAM usage">
    } ]
}
```
