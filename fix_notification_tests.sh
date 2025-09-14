#!/bin/bash

# Fix notification service tests - apply response parsing fixes

echo "Fixing notification service tests..."

# Fix UpdateNotification test
sed -i '' 's/var response models\.Notification/var responseWrapper struct {\
		Notification models.Notification `json:"notification"`\
	}/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/err = json\.Unmarshal(w\.Body\.Bytes(), &response)/err = json.Unmarshal(w.Body.Bytes(), \&responseWrapper)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, updateData\.Status, response\.Status)/assert.Equal(t, updateData.Status, responseWrapper.Notification.Status)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, updateData\.Priority, response\.Priority)/assert.Equal(t, updateData.Priority, responseWrapper.Notification.Priority)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, updateData\.Subject, response\.Subject)/assert.Equal(t, updateData.Subject, responseWrapper.Notification.Subject)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, updateData\.Content, response\.Content)/assert.Equal(t, updateData.Content, responseWrapper.Notification.Content)/' microservices/notification-service/tests/unit/notification_test.go

# Fix SendNotification test
sed -i '' 's/var response map\[string\]interface{}/var responseWrapper struct {\
		Message string `json:"message"`\
	}/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/err = json\.Unmarshal(w\.Body\.Bytes(), &response)/err = json.Unmarshal(w.Body.Bytes(), \&responseWrapper)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, "sent", response\["status"\])/assert.Equal(t, "Notification sent successfully", responseWrapper.Message)/' microservices/notification-service/tests/unit/notification_test.go

# Fix CreateTemplate test
sed -i '' 's/var response models\.Template/var responseWrapper struct {\
		Template models.Template `json:"template"`\
	}/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/err = json\.Unmarshal(w\.Body\.Bytes(), &response)/err = json.Unmarshal(w.Body.Bytes(), \&responseWrapper)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, template\.Name, response\.Name)/assert.Equal(t, template.Name, responseWrapper.Template.Name)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, template\.Description, response\.Description)/assert.Equal(t, template.Description, responseWrapper.Template.Description)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, template\.Type, response\.Type)/assert.Equal(t, template.Type, responseWrapper.Template.Type)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, template\.Category, response\.Category)/assert.Equal(t, template.Category, responseWrapper.Template.Category)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, template\.Subject, response\.Subject)/assert.Equal(t, template.Subject, responseWrapper.Template.Subject)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, template\.Content, response\.Content)/assert.Equal(t, template.Content, responseWrapper.Template.Content)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, template\.IsActive, response\.IsActive)/assert.Equal(t, template.IsActive, responseWrapper.Template.IsActive)/' microservices/notification-service/tests/unit/notification_test.go

# Fix CreateChannel test
sed -i '' 's/var response models\.Channel/var responseWrapper struct {\
		Channel models.Channel `json:"channel"`\
	}/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/err = json\.Unmarshal(w\.Body\.Bytes(), &response)/err = json.Unmarshal(w.Body.Bytes(), \&responseWrapper)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, channel\.Name, response\.Name)/assert.Equal(t, channel.Name, responseWrapper.Channel.Name)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, channel\.Description, response\.Description)/assert.Equal(t, channel.Description, responseWrapper.Channel.Description)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, channel\.Type, response\.Type)/assert.Equal(t, channel.Type, responseWrapper.Channel.Type)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, channel\.Provider, response\.Provider)/assert.Equal(t, channel.Provider, responseWrapper.Channel.Provider)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, channel\.Config, response\.Config)/assert.Equal(t, channel.Config, responseWrapper.Channel.Config)/' microservices/notification-service/tests/unit/notification_test.go

sed -i '' 's/assert\.Equal(t, channel\.IsActive, response\.IsActive)/assert.Equal(t, channel.IsActive, responseWrapper.Channel.IsActive)/' microservices/notification-service/tests/unit/notification_test.go

echo "Notification service tests fixed!"
