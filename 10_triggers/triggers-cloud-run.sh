gsutil mb gs://aerobic-botany-270918-input  
gcloud alpha events triggers create face-finder \
--target-service pigo-events \
--type com.google.cloud.auditlog.event \
--parameters methodName=storage.buckets.update,serviceName=storage.googleapis.com,resourceName=projects/_/buckets/aerobic-botany-270918-input
