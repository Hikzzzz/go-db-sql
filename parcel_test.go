package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("postgres", "user=username dbname=mydb sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.AddParcel(parcel)
	require.NoError(t, err)
	require.Greater(t, id, 0)

	parcel.Number = id // обновляем поле Number у parcel

	storedParcel, err := store.GetParcel(id)
	require.NoError(t, err)
	assert.Equal(t, parcel, storedParcel)

	err = store.DeleteParcel(id)
	require.NoError(t, err)

	_, err = store.GetParcel(id)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("postgres", "user=username dbname=mydb sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.AddParcel(parcel)
	require.NoError(t, err)
	require.Greater(t, id, 0)

	// update address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// get updated parcel
	storedParcel, err := store.GetParcel(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, storedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("postgres", "user=username dbname=mydb sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.AddParcel(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	newStatus := ParcelStatusDelivered
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	storedParcel, err := store.GetParcel(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, storedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {

	db, err := sql.Open("postgres", "user=username dbname=mydb sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.AddParcel(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetParcelsByClient(client)
	require.NoError(t, err)
	assert.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		expectedParcel, exists := parcelMap[parcel.Number]
		require.True(t, exists)
		require.Equal(t, expectedParcel.Client, parcel.Client)
		require.Equal(t, expectedParcel.Status, parcel.Status)
		require.Equal(t, expectedParcel.Address, parcel.Address)
	}
}
