### Fungsi `InitDB()`

Menginisialisasi basis data dengan nama `file.db`. Fungsi ini membuat bucket `Tasks` dan `Categories` jika belum ada. Mengembalikan pointer ke objek `Data` yang berisi koneksi ke basis data jika berhasil, dan error jika gagal.

### Fungsi `(data *Data) StoreTask(task model.Task)`

Menyimpan tugas ke dalam basis data. Mengembalikan error jika terjadi masalah saat menyimpan.

### Fungsi `(data *Data) StoreCategory(category model.Category)`

Menyimpan kategori ke dalam basis data. Mengembalikan error jika terjadi masalah saat menyimpan.

### Fungsi `(data *Data) UpdateTask(id int, task model.Task)`

Memperbarui tugas yang sudah ada berdasarkan `id`. Menggunakan fungsi `StoreTask` untuk menyimpan perubahan.

### Fungsi `(data *Data) UpdateCategory(id int, category model.Category)`

Memperbarui kategori yang sudah ada berdasarkan `id`. Menggunakan fungsi `StoreCategory` untuk menyimpan perubahan.

### Fungsi `(data *Data) DeleteTask(id int)`

Menghapus tugas berdasarkan `id`. Mengembalikan error jika terjadi masalah saat penghapusan.

### Fungsi `(data *Data) DeleteCategory(id int)`

Menghapus kategori berdasarkan `id`. Mengembalikan error jika terjadi masalah saat penghapusan.

### Fungsi `(data *Data) GetTaskByID(id int)`

Mengambil tugas berdasarkan `id`. Mengembalikan objek `model.Task` jika berhasil dan error jika tugas tidak ditemukan atau terjadi masalah lain.

### Fungsi `(data *Data) GetCategoryByID(id int)`

Mengambil kategori berdasarkan `id`. Mengembalikan objek `model.Category` jika berhasil dan error jika kategori tidak ditemukan atau terjadi masalah lain.

### Fungsi `(data *Data) GetTasks()`

Mengambil semua tugas dari basis data. Mengembalikan slice dari `model.Task` jika berhasil dan error jika terjadi masalah.

### Fungsi `(data *Data) GetCategories()`

Mengambil semua kategori dari basis data. Mengembalikan slice dari `model.Category` jika berhasil dan error jika terjadi masalah.

### Fungsi `(data *Data) Reset()`

Menghapus semua bucket (`Tasks` dan `Categories`) dan membuatnya kembali. Mengembalikan error jika terjadi masalah saat penghapusan atau pembuatan bucket.

### Fungsi `(data *Data) CloseDB()`

Menutup koneksi ke basis data. Mengembalikan error jika terjadi masalah saat penutupan.

### Fungsi `(data *Data) GetTaskListByCategory(categoryID int)`

Mengambil daftar tugas yang terkait dengan kategori tertentu. Mengembalikan slice dari `model.TaskCategory` jika berhasil dan error jika kategori tidak ditemukan atau terjadi masalah lain.
