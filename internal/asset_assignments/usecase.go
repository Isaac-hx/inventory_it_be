package assetassignments

import (
	"context"
	"database/sql"
	"errors"
	"inventory-it/internal/assets"
	"inventory-it/internal/middleware"
	"inventory-it/internal/pkg"
	"inventory-it/internal/user"
	"log"
	"time"

	"github.com/google/uuid"
)

type UsecaseAssetAssingments interface {
	GetAllAssetAssignments(ctx context.Context, filter AssetAssignmentFilter) ([]AssetAssignment, pkg.PaginationMeta, error)
	GetAssetAssigmentById(ctx context.Context, assetAssignmentId string) (AssetAssignment, error)
	CreateAssignment(ctx context.Context, assetAssignment AssetAssignment) error
	UpdateAssignmentById(ctx context.Context, assignmentId string, updateAssignment AssetAssignment) error
	GetAllAssignmentsData(ctx context.Context) ([]AssetAssignment, error)
	UpdateAssignmentStatus(ctx context.Context, assignmentId string, statusAssignment AssignmentStatus) error
	GetAssetAssignmentByUserId(ctx context.Context, status AssignmentStatus) ([]AssetAssignment, error)
}

type usecase struct {
	db        *sql.DB
	repo      Repository
	assetRepo assets.Repository
	userRepo  user.Repository
}

func NewUsecaseAssetAssignment(db *sql.DB, repo Repository, assetRepo assets.Repository, userRepo user.Repository) UsecaseAssetAssingments {
	return &usecase{
		db:        db,
		repo:      repo,
		assetRepo: assetRepo,
		userRepo:  userRepo,
	}
}

func (u *usecase) GetAllAssetAssignments(ctx context.Context, filter AssetAssignmentFilter) ([]AssetAssignment, pkg.PaginationMeta, error) {
	assignments, err := u.repo.GetAllAssetAssignments(ctx, filter)
	if err != nil {
		return nil, pkg.PaginationMeta{}, err
	}

	meta, err := u.repo.GetTotalPageAndTotalDataAssetAssignments(ctx, filter)
	if err != nil {
		return nil, pkg.PaginationMeta{}, err
	}
	return assignments, meta, nil
}

func (u *usecase) GetAssetAssigmentById(ctx context.Context, assetAssignmentId string) (AssetAssignment, error) {
	assignment, err := u.repo.GetAssetAssignmentById(ctx, assetAssignmentId)
	if err != nil {
		return AssetAssignment{}, err
	}
	return assignment, nil
}

func (u *usecase) CreateAssignment(ctx context.Context, assetAssignment AssetAssignment) error {
	//	1. Ambil dan validasi claims dari context di awal sebelum memulai transaksi database
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Langsung defer rollback setelah tx berhasil dibuat
	defer tx.Rollback()
	rawClaims := ctx.Value(middleware.Claimskey)
	if rawClaims == nil {
		return errors.New("unauthorized: user data is not valid")
	}

	claim, ok := rawClaims.(*pkg.Claims)
	if !ok {
		return errors.New("unauthorized: format claims is not valid")
	}

	// 2. Get data user dan pastikan tidak nil
	userData, err := u.userRepo.GetUserById(ctx, assetAssignment.UserId)
	if err != nil {
		return err
	}
	//

	// 3. Get data asset dan pastikan tidak nil
	assetData, err := u.assetRepo.GetAssetById(ctx, assetAssignment.AssetId)
	if err != nil {
		return err
	}

	//check status_assignment
	switch assetAssignment.Status {
	case Assigned:
		err := u.assetRepo.UpdateAssetStatusById(ctx, tx, assetData.AssetId, assets.Assigned)
		if err != nil {
			return err
		}
	case Returned:
		// Jika dikembalikan normal, aset menjadi tersedia kembali (Avail)
		err := u.assetRepo.UpdateAssetStatusById(ctx, tx, assetData.AssetId, assets.Available)
		if err != nil {
			return err
		}

	case Lost:
		// Jika hilang, status aset diubah menjadi lost
		err := u.assetRepo.UpdateAssetStatusById(ctx, tx, assetData.AssetId, assets.Retired)
		if err != nil {
			return err
		}
	}

	// 4. Mulai transaksi database tepat sebelum operasi write dilakukan

	// 5. Mapping data dengan benar ke variabel baru
	assetAssignmentData := AssetAssignment{
		AssignmentId: uuid.NewString(),
		AssetId:      assetData.AssetId, // Memakai data dari assetData yang sudah valid
		UserId:       userData.UserId,   // Memakai data dari userData yang sudah valid
		AssignedById: claim.UserID,
		AssignedDate: assetAssignment.AssignedDate,
		Status:       assetAssignment.Status,
		Notes:        assetAssignment.Notes,
		Corporation:  assetAssignment.Corporation,
	}

	// 6. Eksekusi query ke repository
	err = u.repo.CreateAssignmentTx(ctx, tx, assetAssignmentData)
	if err != nil {
		return err
	}

	// 7. Commit transaksi
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
func (u *usecase) UpdateAssignmentById(ctx context.Context, assignmentId string, updateAssignment AssetAssignment) error {
	// 1. Ambil data assignment lama terlebih dahulu untuk validasi & mendapatkan data AssetId/UserId asli
	// (Ini penting agar tidak asal menimpa data jika input dari user kurang lengkap)
	currentAssignment, err := u.repo.GetAssetAssignmentById(ctx, assignmentId)
	if err != nil {
		return err
	}

	// 2. Memulai Transaksi Database
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Pastikan rollback dijalankan jika terjadi error sebelum Commit
	defer tx.Rollback()

	// 3. Destructuring & Mapping Data Baru
	var updatedAssetAssignmentData AssetAssignment
	updatedAssetAssignmentData.AssetId = currentAssignment.AssetId
	updatedAssetAssignmentData.UserId = currentAssignment.UserId
	updatedAssetAssignmentData.Notes = updateAssignment.Notes
	updatedAssetAssignmentData.Status = updateAssignment.Status
	updatedAssetAssignmentData.Corporation = updateAssignment.Corporation
	updatedAssetAssignmentData.ReturnDate = currentAssignment.ReturnDate // Default pakai data lama

	// 4. Logika Perubahan Status Menggunakan Switch Case
	var targetAssetStatus assets.AssetStatus

	switch updateAssignment.Status {
	case Returned:
		targetAssetStatus = assets.Available // Kembali normal -> Berubah jadi avail
		now := time.Now()
		updatedAssetAssignmentData.ReturnDate = &now

	case Damaged:
		targetAssetStatus = assets.Maintenance // Kembali rusak -> Berubah jadi maintenance
		now := time.Now()
		updatedAssetAssignmentData.ReturnDate = &now

	case Lost:
		targetAssetStatus = assets.Retired // Hilang -> Berubah jadi lost
		now := time.Now()
		updatedAssetAssignmentData.ReturnDate = &now

	default:
		// Jika status tidak berubah atau di luar 3 di atas, tidak perlu mengubah status master asset
		targetAssetStatus = ""
	}

	// 5. Eksekusi Perubahan ke Database (Wajib Menggunakan TX / Transaksi)
	if targetAssetStatus != "" {
		// Pastikan method repositori ini mendukung parameter 'tx' agar sinkron dengan transaksi utama
		err := u.assetRepo.UpdateAssetStatusById(ctx, tx, currentAssignment.AssetId, assets.AssetStatus(targetAssetStatus))
		if err != nil {
			return err
		}
	}

	// 6. Update Data Transaksi Assignment
	err = u.repo.UpdateAssetAssignmentByIdTx(ctx, tx, assignmentId, updatedAssetAssignmentData)
	if err != nil {
		return err
	}

	// 7. Commit Transaksi jika semua sukses
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (u *usecase) GetAllAssignmentsData(ctx context.Context) ([]AssetAssignment, error) {
	return u.repo.GetAllAssignmentsData(ctx)
}

func (u *usecase) UpdateAssignmentStatus(ctx context.Context, assignmentId string, statusAssignment AssignmentStatus) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	//get assignment id
	assignment, err := u.repo.GetAssetAssignmentById(ctx, assignmentId)
	if err != nil {
		return err
	}
	log.Println(assignment.Status)
	//update status asset
	err = u.assetRepo.UpdateAssetStatusById(ctx, tx, assignment.AssetId, assets.Available)
	if err != nil {
		return err
	}
	now := time.Now()
	var updatedStatusAssignment AssetAssignment
	updatedStatusAssignment.ReturnDate = &now
	updatedStatusAssignment.Status = statusAssignment
	//update status assignment
	err = u.repo.UpdateStatusAssignmentByIdTx(ctx, tx, assignmentId, updatedStatusAssignment)
	if err != nil {
		return err
	}
	// 5. COMMIT TRANSAKSI (Ini yang bikin data tersimpan permanen)
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (u *usecase) GetAssetAssignmentByUserId(ctx context.Context, status AssignmentStatus) ([]AssetAssignment, error) {
	rawClaims := ctx.Value(middleware.Claimskey)
	if rawClaims == nil {
		return nil, errors.New("unauthorized: user data is not valid")
	}

	claim, ok := rawClaims.(*pkg.Claims)
	if !ok {
		return nil, errors.New("unauthorized: format claims is not valid")
	}

	return u.repo.GetAssignmentsByUserId(ctx, claim.UserID, status)
}
