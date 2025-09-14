#!/bin/bash

# Fix authentication middleware in component service tests
sed -i '' 's/router := gin\.New()/router := gin.New()\
	router.Use(func(c *gin.Context) {\
		c.Set("tenant_id", uint(1))\
		c.Next()\
	})/g' microservices/component-service/tests/unit/component_test.go

# Fix response parsing for wrapped responses
sed -i '' 's/var response models\.Component/var responseWrapper struct {\
		Component models.Component `json:"component"`\
	}/g' microservices/component-service/tests/unit/component_test.go

sed -i '' 's/err = json\.Unmarshal(w\.Body\.Bytes(), &response)/err = json.Unmarshal(w.Body.Bytes(), \&responseWrapper)/g' microservices/component-service/tests/unit/component_test.go

sed -i '' 's/assert\.Equal(t, component\.ID, response\.ID)/assert.Equal(t, component.ID, responseWrapper.Component.ID)/g' microservices/component-service/tests/unit/component_test.go

sed -i '' 's/assert\.Equal(t, component\.Name, response\.Name)/assert.Equal(t, component.Name, responseWrapper.Component.Name)/g' microservices/component-service/tests/unit/component_test.go

echo "Component service tests fixed!"
