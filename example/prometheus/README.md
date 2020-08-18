# Prometheus Examples

SanePA exposes some metrics that Prometheus can pull.

## Scaling Alerts

SanePA is scaling up: `increase(sanepa_scale_up_events[5m]) > 0`

SanePA is scaling down: `increase(sanepa_scale_down_events[5m]) > 0`

## SanePA Metrics Alerts

SanePA is failing to pull metrics from the Kubernetes API: `increase(sanepa_metric_collection_errors[5m]) > 0`