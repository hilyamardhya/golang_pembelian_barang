using Api_Sddp.MODELS;
using System.Data;
using Microsoft.AspNetCore.Mvc;
using MySql.Data.MySqlClient;
using Dapper;


namespace API_Beli_barang.Controllers
{
    [Route("api/")] // jalur dari API ke database
    [ApiController] // untuk mengontrol
    public class barangController : Controller //constructure
    {
        private readonly IConfiguration _config;

        public barangController(IConfiguration config)
        {
            _config = config;
        }

        private IDbConnection Connection => new MySqlConnection(_config.GetConnectionString("Default"));

        [HttpGet]
        public ActionResult<IEnumerable<barang>> GetAll()
        {
            using (var conn = Connection)
            {
                var barang = conn.Query<barang>("SELECT * FROM barang");
                return Ok(barang);
            }
        }
        [HttpPost("pembelian")]
        public IActionResult PostPembelian([FromBody] pembelian pembelian)
        {
            using (var conn = Connection)
            {
                // Cari barang berdasarkan NamaBarang
                var barang = conn.QueryFirstOrDefault<barang>("SELECT * FROM Barang WHERE NamaBarang = @NamaBarang", new { NamaBarang = pembelian.Barang });

                if (barang == null)
                {
                    return BadRequest(new { message = "Barang tidak ditemukan." });
                }

                // Validasi stok
                if (barang.Stok < pembelian.Jumlah)
                {
                    return BadRequest(new { message = "Stok tidak mencukupi." });
                }

                // Simpan pembelian
                var sql = "INSERT INTO Pembelian (Nama, Barang, Jumlah, TotalBayar, IdBarang) VALUES (@Nama, @Barang, @Jumlah, @TotalBayar, @IdBarang)";
                pembelian.TotalBayar = pembelian.Jumlah * (double)barang.Harga; // Casting Harga (decimal) ke double
                pembelian.IdBarang = barang.IdBarang;
                var result = conn.Execute(sql, pembelian);

                return Ok(new { message = "Pembelian berhasil disimpan." });
            }
        }
        [HttpPut("updateStok/{IdBarang}")]
        public ActionResult UpdateStok(int IdBarang, [FromBody] int jumlahPembeli)
        {
            using (var conn = Connection)
            {
                // Cek apakah barang ada di database
                var barang = conn.QueryFirstOrDefault<barang>("SELECT * FROM barang WHERE IdBarang = @IdBarang", new { IdBarang = IdBarang });
                if (barang == null)
                {
                    return NotFound("barang tidak ditemukan");
                }

                // Pastikan stok cukup untuk Pembeli
                if (barang.Stok < jumlahPembeli)
                {
                    return BadRequest("Stok tidak cukup");
                }

                // Mengurangi stok barang
                var newStok = barang.Stok - jumlahPembeli;
                var updateQuery = "UPDATE barang SET Stok = @Stok WHERE IdBarang = @IdBarang";
                var result = conn.Execute(updateQuery, new { Stok = newStok, IdBarang = IdBarang });

                if (result == 0)
                {
                    return StatusCode(500, "Gagal mengupdate stok");
                }

                return Ok("Stok berhasil diperbarui");
            }
        }

        [HttpGet("riwayatPembeli/{namaPembeli}")]
        public ActionResult<IEnumerable<pembelian>> GetRiwayatPembeli(string namaPembeli)
        {
            using (var conn = Connection)
            {
                var sql = "SELECT * FROM pembelian WHERE Nama = @NamaPembeli";
                var riwayat = conn.Query<pembelian>(sql, new { NamaPembeli = namaPembeli });

                if (riwayat == null || !riwayat.Any())
                {
                    return NotFound(new { message = "Tidak ada riwayat Pembeli untuk nama tersebut." });
                }

                return Ok(riwayat);
            }
        }


    }
}