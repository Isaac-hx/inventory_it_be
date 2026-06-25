package maintenances

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"inventory-it/internal/assets"
	"inventory-it/internal/pkg"
	"time"

	"github.com/google/uuid"
)

type Usecase interface {
	GetAllMaintenances(context.Context, MaintenanceFilter) ([]Maintenance, pkg.PaginationMeta, error)
	GetAllMaintenancesData(context.Context) ([]Maintenance, error)
	GetMaintenanceById(context.Context, string) (Maintenance, error)
	CreateMaintenance(context.Context, Maintenance) (Maintenance, error)
	UpdateMaintenance(context.Context, string, Maintenance) error
	UpdateStatusMaintenance(context.Context, string, string, *time.Time) error
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
	maintenanceData.Status = maintenance.Status
	maintenanceData.AssetId = maintenance.AssetId
	maintenanceData.MaintenanceAt = maintenance.MaintenanceAt
	maintenanceData.Asset = asset

	//create maintenance with transaction
	err = u.repo.CreateMaintenanceTx(ctx, tx, maintenanceData)
	if err != nil {
		return Maintenance{}, err
	}

	//call repo update asset status with transaction
	err = u.assetRepo.UpdateAssetStatusById(ctx, tx, maintenance.AssetId, assets.Maintenance)
	if err != nil {
		return Maintenance{}, err
	}

	//commit transaction
	if err := tx.Commit(); err != nil {
		return Maintenance{}, err
	}

	return maintenanceData, nil

}

func (u *usecase) GetAllMaintenances(ctx context.Context, maintenanceFilter MaintenanceFilter) ([]Maintenance, pkg.PaginationMeta, error) {

	maintenances, err := u.repo.GetAllMaintenances(ctx, maintenanceFilter)
	if err != nil {
		return nil, pkg.PaginationMeta{}, err

	}
	meta, err := u.repo.GetTotalPageAndTotalDataMaintenances(ctx, maintenanceFilter)
	if err != nil {
		return nil, pkg.PaginationMeta{}, err

	}

	return maintenances, meta, nil

}

func (u *usecase) GetMaintenanceById(ctx context.Context, maintenanceId string) (Maintenance, error) {
	maintenance, err := u.repo.GetMaintenanceById(ctx, maintenanceId)
	if err != nil {
		return Maintenance{}, err
	}

	return maintenance, nil
}

func (u *usecase) UpdateMaintenance(ctx context.Context, maintenance_id string, maintenance Maintenance) error {
	// 1. Mulai transaksi database
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Otomatis rollback jika terjadi kegagalan di tengah jalan
	defer tx.Rollback()

	// 2. Ambil data maintenance yang sudah ada (Pastikan repo ini mendukung transaksi jika diperlukan)
	existingMaintenance, err := u.repo.GetMaintenanceById(ctx, maintenance_id)
	if err != nil {
		return err
	}

	// 3. Salin data lama ke objek baru untuk melacak perubahan
	updatedMaintenance := existingMaintenance

	// 4. Perbarui data dasar jika dikirim dari payload frontend
	if maintenance.Description != "" {
		updatedMaintenance.Description = maintenance.Description
	}
	if maintenance.Cost >= 0 {
		updatedMaintenance.Cost = maintenance.Cost
	}

	// 5. Logika Penggabungan Status & Waktu Penyelesaian (CompletedAt)
	if maintenance.Status != "" {
		updatedMaintenance.Status = maintenance.Status

		if maintenance.Status == Completed || maintenance.Status == Cancelled {
			now := time.Now()
			updatedMaintenance.CompletedAt = &now
		} else {
			updatedMaintenance.CompletedAt = nil
		}
	}

	// 6. Eksekusi pembaruan data Log Maintenance di dalam Transaksi
	// Disarankan membuat satu fungsi repo general untuk update semua field sekaligus agar hemat query
	err = u.repo.UpdateMaintenanceTx(ctx, tx, maintenance_id, updatedMaintenance)
	if err != nil {
		return errors.New("failed to update maintenance record in transaction")
	}

	// 7. Logika Efek Domino: Update Status Aset di IT Inventory
	if maintenance.Status != "" {
		var checkStatus assets.AssetStatus

		if updatedMaintenance.Status == Completed || updatedMaintenance.Status == Cancelled {
			checkStatus = assets.Available
		} else {
			checkStatus = assets.Maintenance
		}

		// Jalankan pembaruan status aset terikat menggunakan context transaksi (tx) yang sama
		err = u.assetRepo.UpdateAssetStatusById(ctx, tx, updatedMaintenance.AssetId, checkStatus)
		if err != nil {
			return errors.New("failed to cascade update asset status by id")
		}
	}

	// 8. Komit transaksi jika seluruh operasi di atas sukses tanpa hambatan
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (u *usecase) GetAllMaintenancesData(ctx context.Context) ([]Maintenance, error) {
	return u.repo.GetAllMaintenancesData(ctx)
}

func (u *usecase) UpdateStatusMaintenance(ctx context.Context, maintenanceId string, status string, completedAt *time.Time) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()
	var updateMaintenanceData Maintenance
	updateMaintenanceData.Status = MaintenanceStatus(status)
	updateMaintenanceData.CompletedAt = completedAt
	maintenance, err := u.repo.GetMaintenanceById(ctx, maintenanceId)
	if err != nil {
		return err
	}

	if updateMaintenanceData.Status == Completed || updateMaintenanceData.Status == Cancelled {

		//get asset by id and update
		err = u.assetRepo.UpdateAssetStatusById(ctx, tx, maintenance.AssetId, assets.Available)
		if err != nil {
			return err
		}
	} else {
		err = u.assetRepo.UpdateAssetStatusById(ctx, tx, maintenance.AssetId, assets.Maintenance)
	}

	err = u.repo.UpdateMaintenanceStatusTx(ctx, tx, maintenanceId, updateMaintenanceData)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
