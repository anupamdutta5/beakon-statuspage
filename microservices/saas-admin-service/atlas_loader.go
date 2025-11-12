// +build tools

package main

import (
	"ariga.io/atlas-provider-gorm/gormschema"
	"github.com/anupamdutta5/saas-admin-service/internal/models"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		&models.Platform{},
		&models.SaaSPlan{},
		&models.PricingTier{},
		&models.PricingFeature{},
		&models.PlanFeature{},
		&models.SaaSFeature{},
		&models.SaaSFeatureFlag{},
		&models.SaaSAdminUser{},
		&models.SaaSTenant{},
		&models.SaaSNotification{},
		&models.SaaSActivity{},
		&models.SaaSBackup{},
		&models.SaaSStats{},
		&models.SaaSIncident{},
		&models.SaaSIncidentUpdate{},
		&models.SaaSMaintenanceWindow{},
		&models.SaaSService{},
		&models.SaaSIntegration{},
		&models.SaaSWebhook{},
		&models.SaaSSubscriber{},
		&models.SaaSComponent{},
		&models.CustomDomain{},
		&models.Session{},
	)
	if err != nil {
		panic(err)
	}
	for _, stmt := range stmts {
		println(stmt)
	}
}