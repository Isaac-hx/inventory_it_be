package maintenances

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	assetassignments "inventory-it/internal/asset_assignments"
	"inventory-it/internal/assets"
	"inventory-it/internal/middleware"
	"inventory-it/internal/pkg"

	"github.com/google/uuid"
)

type Usecase interface {
	GetAllMaintenances(context.Context, MaintenanceFilter) ([]Maintenance, pkg.PaginationMeta, error)
	GetAllMaintenancesData(context.Context) ([]Maintenance, error)
	GetMaintenanceById(context.Context, string) (Maintenance, error)
	CreateMaintenance(context.Context, Maintenance) (Maintenance, error)
	UpdateMaintenance(context.Context, string, Maintenance) error
	UpdateStatusMaintenance(context.Context, string, string, *time.Time) error
	CreateRequest(context.Context, Maintenance) error
	GetAllMaintenancesByUserId(context.Context) ([]Maintenance, error)
}

type usecase struct {
	db                   *sql.DB
	repo                 Repository
	assetRepo            assets.Repository
	assetassignmentsRepo assetassignments.Repository
}

func NewMaintenanceUsecase(
	db *sql.DB,
	repo Repository,
	assetRepo assets.Repository,
	assignmentRepo assetassignments.Repository,
) Usecase {
	return &usecase{
		db:                   db,
		repo:                 repo,
		assetRepo:            assetRepo,
		assetassignmentsRepo: assignmentRepo,
	}
}

func (u *usecase) CreateMaintenance(ctx context.Context, maintenance Maintenance) (Maintenance, error) {
	// 1. Mulai transaksi database
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return Maintenance{}, err
	}
	defer tx.Rollback()

	// 2. Ambil data Assignment berdasarkan AssignmentId, bukan lagi langsung via AssetId
	assignment, err := u.assetassignmentsRepo.GetAssetAssignmentById(ctx, maintenance.Assignment.AssignmentId)
	if err != nil {
		return Maintenance{}, fmt.Errorf("failed to retrieve asset assignment: %w", err)
	}

	// 3. Ambil data aset dari relasi assignment
	asset, err := u.assetRepo.GetAssetById(ctx, assignment.Asset.AssetId)
	if err != nil {
		return Maintenance{}, err
	}

	// 4. Pastikan status aset saat ini valid (Available atau Assigned) untuk bisa dimaintenance
	if asset.Status == assets.Maintenance {
		return Maintenance{}, fmt.Errorf("asset with id %s is already undergoing maintenance", asset.AssetId)
	}

	// 5. Konstruksi data Maintenance baru dengan mereferensikan AssignmentId
	var maintenanceData Maintenance
	maintenanceData.MaintenanceId = uuid.New().String()
	maintenanceData.Description = maintenance.Description
	maintenanceData.Cost = maintenance.Cost
	maintenanceData.Status = maintenance.Status
	maintenanceData.MaintenanceAt = maintenance.MaintenanceAt
	maintenanceData.Assignment.AssignmentId = assignment.AssignmentId

	// 6. Jalankan pembuatan maintenance di dalam transaksi
	err = u.repo.CreateMaintenanceTx(ctx, tx, maintenanceData)
	if err != nil {
		return Maintenance{}, err
	}

	// 7. Update status aset menjadi 'Maintenance'
	err = u.assetRepo.UpdateAssetStatusById(ctx, tx, asset.AssetId, assets.Maintenance)
	if err != nil {
		return Maintenance{}, err
	}

	// 8. Commit transaksi database
	if err := tx.Commit(); err != nil {
		return Maintenance{}, err
	}

	maintenanceData.Asset = asset
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

	return listCompletedCheck(maintenances), meta, nil
}

func (u *usecase) GetMaintenanceById(ctx context.Context, maintenanceId string) (Maintenance, error) {
	maintenance, err := u.repo.GetMaintenanceById(ctx, maintenanceId)
	if err != nil {
		return Maintenance{}, err
	}
	return maintenance, nil
}

func (u *usecase) UpdateMaintenance(ctx context.Context, maintenanceId string, maintenance Maintenance) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	existingMaintenance, err := u.repo.GetMaintenanceById(ctx, maintenanceId)
	if err != nil {
		return err
	}

	updatedMaintenance := existingMaintenance

	if maintenance.Description != "" {
		updatedMaintenance.Description = maintenance.Description
	}
	if maintenance.Cost >= 0 {
		updatedMaintenance.Cost = maintenance.Cost
	}

	// Logika transisi status & waktu penyelesaian
	if maintenance.Status != "" {
		updatedMaintenance.Status = maintenance.Status
		if maintenance.Status == Completed || maintenance.Status == Cancelled {
			now := time.Now()
			updatedMaintenance.CompletedAt = &now
		} else {
			updatedMaintenance.CompletedAt = nil
		}
	}

	err = u.repo.UpdateMaintenanceTx(ctx, tx, maintenanceId, updatedMaintenance)
	if err != nil {
		return errors.New("failed to update maintenance record in transaction")
	}

	// Efek cascade update status aset berdasarkan status maintenance yang diperbarui
	if maintenance.Status != "" {
		var nextAssetStatus assets.AssetStatus
		if updatedMaintenance.Status == Completed || updatedMaintenance.Status == Cancelled {
			// Kembalikan ke status operasional semula (Assigned/Available)
			nextAssetStatus = assets.Available
		} else {
			nextAssetStatus = assets.Maintenance
		}

		err = u.assetRepo.UpdateAssetStatusById(ctx, tx, updatedMaintenance.Assignment.AssetId, nextAssetStatus)
		if err != nil {
			return errors.New("failed to cascade update asset status by id")
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (u *usecase) GetAllMaintenancesData(ctx context.Context) ([]Maintenance, error) {
	maintenances, err := u.repo.GetAllMaintenancesData(ctx)
	if err != nil {
		return nil, err
	}
	return listCompletedCheck(maintenances), nil
}

func (u *usecase) UpdateStatusMaintenance(ctx context.Context, maintenanceId string, status string, completedAt *time.Time) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	maintenance, err := u.repo.GetMaintenanceById(ctx, maintenanceId)
	if err != nil {
		return err
	}

	var updateMaintenanceData Maintenance
	updateMaintenanceData.Status = MaintenanceStatus(status)
	updateMaintenanceData.CompletedAt = completedAt

	assignment, err := u.assetassignmentsRepo.GetAssetAssignmentById(ctx, maintenance.Assignment.AssignmentId)
	// Manajemen status aset ketika status maintenance diperbarui
	switch updateMaintenanceData.Status {
	case Pending:
		err = u.assetRepo.UpdateAssetStatusById(ctx, tx, assignment.AssetId, assets.Assigned)
	case InProgress:
		err = u.assetRepo.UpdateAssetStatusById(ctx, tx, assignment.AssetId, assets.Maintenance)
	case Completed, Cancelled:
		err = u.assetRepo.UpdateAssetStatusById(ctx, tx, assignment.AssetId, assets.Assigned)
	}

	if err != nil {
		return fmt.Errorf("failed to update cascade asset status: %w", err)
	}

	err = u.repo.UpdateMaintenanceTx(ctx, tx, maintenanceId, updateMaintenanceData)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (u *usecase) CreateRequest(ctx context.Context, createMaintenance Maintenance) error {
	rawClaims := ctx.Value(middleware.Claimskey)
	if rawClaims == nil {
		return errors.New("unauthorized: user data is not valid")
	}

	claim, ok := rawClaims.(*pkg.Claims)
	if !ok {
		return errors.New("unauthorized: format claims is not valid")
	}

	// Validasi bahwa AssignmentId yang dikirimkan user memang valid
	assignment, err := u.assetassignmentsRepo.GetAssetAssignmentById(ctx, createMaintenance.Assignment.AssignmentId)
	if err != nil {
		return fmt.Errorf("invalid asset assignment: %w", err)
	}

	var maintenanceRequest Maintenance
	maintenanceRequest.MaintenanceId = uuid.New().String()
	maintenanceRequest.User.UserId = claim.UserID
	maintenanceRequest.Description = createMaintenance.Description
	maintenanceRequest.Assignment.AssetId = assignment.Asset.AssetId // Petakan asset id dari data assignment
	maintenanceRequest.Assignment.AssignmentId = assignment.AssignmentId

	err = u.repo.CreateRequest(ctx, maintenanceRequest)
	if err != nil {
		return err
	}

	return nil
}

func (u *usecase) GetAllMaintenancesByUserId(ctx context.Context) ([]Maintenance, error) {
	rawClaims := ctx.Value(middleware.Claimskey)
	if rawClaims == nil {
		return nil, errors.New("unauthorized: user data is not valid")
	}

	claim, ok := rawClaims.(*pkg.Claims)
	if !ok {
		return nil, errors.New("unauthorized: format claims is not valid")
	}

	maintenances, err := u.repo.GetAllMaintenancesByUserId(ctx, claim.UserID)
	if err != nil {
		return nil, err
	}

	return listCompletedCheck(maintenances), nil
}

// helper mencegah return nil slice ke format response JSON frontend
func listCompletedCheck(m []Maintenance) []Maintenance {
	if m == nil {
		return []Maintenance{}
	}
	return m
}
