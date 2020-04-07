package main

import "time"
type eventProtoPayload struct {
	ProtoPayload struct {
		Type   string `json:"@type"`
		Status struct {
		} `json:"status"`
		AuthenticationInfo struct {
			PrincipalEmail string `json:"principalEmail"`
		} `json:"authenticationInfo"`
		RequestMetadata struct {
			CallerIP                string `json:"callerIp"`
			CallerSuppliedUserAgent string `json:"callerSuppliedUserAgent"`
			RequestAttributes       struct {
				Time time.Time `json:"time"`
				Auth struct {
				} `json:"auth"`
			} `json:"requestAttributes"`
			DestinationAttributes struct {
			} `json:"destinationAttributes"`
		} `json:"requestMetadata"`
		ServiceName       string `json:"serviceName"`
		MethodName        string `json:"methodName"`
		AuthorizationInfo []struct {
			Resource           string `json:"resource"`
			Permission         string `json:"permission"`
			Granted            bool   `json:"granted"`
			ResourceAttributes struct {
			} `json:"resourceAttributes"`
		} `json:"authorizationInfo"`
		ResourceName string `json:"resourceName"`
		ServiceData  struct {
			Type        string `json:"@type"`
			PolicyDelta struct {
				BindingDeltas []struct {
					Action string `json:"action"`
					Role   string `json:"role"`
					Member string `json:"member"`
				} `json:"bindingDeltas"`
			} `json:"policyDelta"`
		} `json:"serviceData"`
		ResourceLocation struct {
			CurrentLocations []string `json:"currentLocations"`
		} `json:"resourceLocation"`
	} `json:"protoPayload"`
	InsertID string `json:"insertId"`
	Resource struct {
		Type   string `json:"type"`
		Labels struct {
			ProjectID  string `json:"project_id"`
			Location   string `json:"location"`
			BucketName string `json:"bucket_name"`
		} `json:"labels"`
	} `json:"resource"`
	Timestamp        time.Time `json:"timestamp"`
	Severity         string    `json:"severity"`
	LogName          string    `json:"logName"`
	ReceiveTimestamp time.Time `json:"receiveTimestamp"`
}