package maintenances

import (
	"context"
	"database/sql"
	"fmt"
	"inventory-it/internal/assets"
	"time"

	"github.com/google/uuid"
)

type Usecase interface {
	GetAllMaintenances(context.Context, MaintenanceFilter) ([]Maintenance, error)
	GetMaintenanceById(context.Context, string) (Maintenance, error)
	CreateMaintenance(context.Context, Maintenance) (Maintenance, error)
	UpdateStatusMaintenance(context.Context, Maintenance) error
	UpdateMaintenance(context.Context, Maintenance) error
}

type usecase struct {
	db        *sql.DB
	repo      Repository
	assetRepo assets.Repository
}

func NewMaintenanceUsecase(db *sql.DB, repo Repository, assetRepo assets.Repository) Usecase {
	return &usecase{
		db:        db,
		repo:      repo,
		assetRepo: assetRepo,
	}
}

func (u *usecase) CreateMaintenance(ctx context.Context, maintenance Maintenance) (Maintenance, error) {

	// Start a transaction
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return Maintenance{}, err
	}
	//if set instruction is error, rollback
	defer tx.Rollback()

	//call search asset by id with transaction
	asset, err := u.assetRepo.GetAssetById(ctx, maintenance.AssetId)
	if err != nil {
		return Maintenance{}, err
	}
	//check asset status is available for maintenance
	if asset.Status != assets.Available {
		return Maintenance{}, fmt.Errorf("asset with id %s is not available for maintenance", maintenance.AssetId)
	}

	//construct maintenance data
	var maintenanceData Maintenance
	maintenanceData.MaintenanceId = uuid.New().String()
	maintenanceData.Description = maintenance.Description
	maintenanceData.Cost = maintenance.Cost
	maintenanceData.Status = Pending
	maintenanceData.AssetId = maintenance.AssetId
	maintenanceData.MaintenanceAt = time.Now()
	maintenanceData.Asset = asset

	//create maintenance with transaction
	err = u.repo.CreateMaintenanceTx(ctx, tx, maintenanceData)
	if err != nil {
		return Maintenance{}, err
	}

	//call repo update asset status with transaction
	err = u.assetRepo.UpdateAssetStatusById(ctx, maintenance.AssetId, assets.Maintenance)
	if err != nil {
		return Maintenance{}, err
	}

	//commit transaction
	if err := tx.Commit(); err != nil {
		return Maintenance{}, err
	}

	return maintenanceData, nil

}

func (u *usecase) GetAllMaintenances(ctx context.Context, maintenanceFilter MaintenanceFilter) ([]Maintenance, error) {
	maintenances, err := u.repo.GetAllMaintenances(ctx, maintenanceFilter)
	if err != nil {
		return nil, err
	}

	return maintenances, nil

}

func (u *usecase) GetMaintenanceById(ctx context.Context, maintenanceId string) (Maintenance, error) {
	maintenance, err := u.repo.GetMaintenanceById(ctx, maintenanceId)
	if err != nil {
		return Maintenance{}, err
	}

	return maintenance, nil
}

func (u *usecase) UpdateStatusMaintenance(ctx context.Context, maintenance Maintenance) error {
	// Start a transaction
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	//if set instruction is error, rollback
	defer tx.Rollback()

	//call search maintenance by id with transaction
	existingMaintenance, err := u.repo.GetMaintenanceById(ctx, maintenance.MaintenanceId)
	if err != nil {
		return err
	}

	var updatedMaintenance Maintenance
	if maintenance.Status == Completed || maintenance.Status == Cancelled {
		updatedMaintenance = existingMaintenance
		updatedMaintenance.Status = maintenance.Status
		updatedMaintenance.CompletedAt = time.Now()
	} else {
		updatedMaintenance = existingMaintenance
		updatedMaintenance.Status = maintenance.Status
		updatedMaintenance.CompletedAt = time.Time{}
	}

	//call repo to update maintenance status with transaction
	err = u.repo.UpdateMaintenanceStatusTx(ctx, tx, updatedMaintenance)
	if err != nil {
		return err
	}

	//check if maintenance status is completed or cancelled, if yes set asset status to available, if not set asset status to maintenance
	var checkStatus assets.AssetStatus
	if maintenance.Status == Completed || maintenance.Status == Cancelled {
		checkStatus = assets.Available
	} else {
		checkStatus = assets.Maintenance
	}

	//call repo asset to update asset status to available
	err = u.assetRepo.UpdateAssetStatusById(ctx, existingMaintenance.AssetId, checkStatus)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (u *usecase) UpdateMaintenance(ctx context.Context, maintenance Maintenance) error {
	// Start a transaction
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	//if set instruction is error, rollback
	defer tx.Rollback()

	//call search maintenance by id with transaction
	existingMaintenance, err := u.repo.GetMaintenanceById(ctx, maintenance.MaintenanceId)
	if err != nil {
		return err
	}

	updatedMaintenance := existingMaintenance
	updatedMaintenance.Description = maintenance.Description
	updatedMaintenance.Cost = maintenance.Cost

	//call repo to update maintenance cost
	err = u.repo.UpdateMaintenanceCostTx(ctx, tx, updatedMaintenance.MaintenanceId, updatedMaintenance.Cost)
	if err != nil {
		return err
	}

	//call repo to update maintenance description
	err = u.repo.UpdateMaintenanceDescriptionTx(ctx, tx, updatedMaintenance.MaintenanceId, updatedMaintenance.Description)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
