package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	notificationProto "github.com/paper-social/notification-service/proto/generated/notification/proto"
	postProto "github.com/paper-social/notification-service/proto/generated/post/proto"
)

type Resolver struct {
	postClient         postProto.PostServiceClient
	notificationClient notificationProto.NotificationServiceClient
}

func NewResolver(notificationClient notificationProto.NotificationServiceClient, postClient postProto.PostServiceClient) *Resolver {
	return &Resolver{notificationClient: notificationClient, postClient: postClient}
}
