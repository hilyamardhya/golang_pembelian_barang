package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const apiURL = "https://localhost:7228/api"

type Barang struct {
	NamaBarang string  `json:"namaBarang"`
	Harga      float64 `json:"harga"`
	Stok       int     `json:"stok"`
	IdBarang   int     `json:"idBarang"`
}

type Pembelian struct {
	Nama       string  `json:"nama"`
	Barang     string  `json:"barang"`
	JumlahBeli int     `json:"jumlah"`
	TotalBayar float64 `json:"totalBayar"`
	IdBarang   int     `json:"idBarang"`
}

type Stack struct {
	elements []Pembelian
}

func (s *Stack) Push(p Pembelian) {
	s.elements = append(s.elements, p)
}

func (s *Stack) Pop() (Pembelian, bool) {
	if len(s.elements) == 0 {
		return Pembelian{}, false
	}
	last := s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]
	return last, true
}

func (s *Stack) Peek() (Pembelian, bool) {
	if len(s.elements) == 0 {
		return Pembelian{}, false
	}
	return s.elements[len(s.elements)-1], true
}

var riwayatPembelian Stack

func main() {
	barang, err := getAllBarang()
	if err != nil {
		log.Fatalf("Gagal mengambil data barang: %v", err)
	}

	displayIntro()
	displayBarangTabel(barang)

	for {
		fmt.Println("\nMenu:")
		fmt.Println("1. Lakukan Pembelian")
		fmt.Println("2. Lihat Riwayat Pembelian")
		fmt.Println("3. Cari Riwayat Pembelian Berdasarkan Nama")
		fmt.Println("4. Keluar")
		fmt.Print("Pilih opsi (1/2/3/4): ")
		var pilihan int
		fmt.Scanln(&pilihan)

		switch pilihan {
		case 1:
			pembelian := inputPembelian(barang)
			if pembelian.TotalBayar > 0 {
				displayTotalBayar(pembelian)
				err := postPembelian(pembelian)
				if err != nil {
					log.Fatalf("Gagal menyimpan pembelian: %v", err)
				}
				err = updateStok(pembelian.Barang, pembelian.JumlahBeli)
				if err != nil {
					log.Fatalf("Gagal memperbarui stok: %v", err)
				}
				riwayatPembelian.Push(pembelian)
			}
		case 2:
			displayRiwayatPembelian()
		case 3:
			fmt.Print("\nMasukkan nama pembeli: ")
			var namaPembeli string
			fmt.Scanln(&namaPembeli)

			riwayat, err := getRiwayatPembeli(namaPembeli)
			if err != nil {
				fmt.Printf("Gagal mendapatkan riwayat pembelian: %v\n", err)
				continue
			}

			if len(riwayat) == 0 {
				fmt.Println("Tidak ada riwayat pembelian untuk nama tersebut.")
			} else {
				fmt.Println("\nRiwayat Pembelian Berdasarkan Nama:")
				for _, p := range riwayat {
					fmt.Printf("Nama: %s, Barang: %s, Jumlah: %d, Total Bayar: %s\n",
						p.Nama, p.Barang, p.JumlahBeli, formatHarga(p.TotalBayar))
				}
			}
		case 4:
			fmt.Println("Terima kasih telah berbelanja di toko kami. Sampai jumpa!")
			return

		default:
			fmt.Println("Pilihan tidak valid, silakan coba lagi.")
		}
	}
}

func displayRiwayatPembelian() {
	if len(riwayatPembelian.elements) == 0 {
		fmt.Println("\nBelum ada riwayat pembelian.")
		return
	}

	fmt.Println("\nRiwayat Pembelian:")
	for i := len(riwayatPembelian.elements) - 1; i >= 0; i-- {
		p := riwayatPembelian.elements[i]
		fmt.Printf("Nama: %s, Barang: %s, Jumlah: %d, Total Bayar: %s\n",
			p.Nama, p.Barang, p.JumlahBeli, formatHarga(p.TotalBayar))
	}
}

func getAllBarang() ([]Barang, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request GET: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengirim request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("gagal mengambil data barang: %s", body)
	}

	var barang []Barang
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons body: %w", err)
	}

	err = json.Unmarshal(body, &barang)
	if err != nil {
		return nil, fmt.Errorf("gagal meng-unmarshal JSON: %w", err)
	}

	return barang, nil
}

func displayIntro() {
	fmt.Println("Selamat datang di Toko Kami!")
	fmt.Println("Kami menyediakan berbagai barang berkualitas.")
	fmt.Println("Berikut adalah daftar barang yang tersedia untuk Anda: ")
	fmt.Println()
}

func formatHarga(amount float64) string {
	return "Rp " + strconv.FormatFloat(amount, 'f', 0, 64)
}

func displayBarangTabel(barang []Barang) {
	fmt.Printf("%-20s %-10s %-10s\n", "Nama Barang", "Harga", "Stok")
	fmt.Println(strings.Repeat("-", 40))

	for _, b := range barang {
		hargaFormatted := formatHarga(b.Harga)
		fmt.Printf("%-20s %-10s %-10d\n", b.NamaBarang, hargaFormatted, b.Stok)
	}
}

func inputPembelian(barang []Barang) Pembelian {
	var pembelian Pembelian
	var barangBeli string
	var jumlahBeli int

	fmt.Print("\nSilakan masukkan nama Anda: ")
	fmt.Scanln(&pembelian.Nama)

	fmt.Print("Pilih barang yang ingin dibeli: ")
	reader := bufio.NewReader(os.Stdin)
	barangBeli, _ = reader.ReadString('\n')
	barangBeli = strings.TrimSpace(barangBeli)

	var productFound bool
	for _, b := range barang {
		if strings.ToLower(b.NamaBarang) == strings.ToLower(barangBeli) {
			productFound = true
			pembelian.Barang = b.NamaBarang
			pembelian.IdBarang = b.IdBarang
			fmt.Print("Masukkan jumlah barang yang ingin dibeli: ")
			fmt.Scanln(&jumlahBeli)

			if jumlahBeli > b.Stok {
				fmt.Println("Stok tidak mencukupi!")
				return Pembelian{}
			}

			pembelian.JumlahBeli = jumlahBeli
			pembelian.TotalBayar = float64(jumlahBeli) * b.Harga
			break
		}
	}

	if !productFound {
		fmt.Println("\nBarang yang Anda masukkan tidak ada di toko kami. Terima kasih.")
		return Pembelian{}
	}

	return pembelian
}

func displayTotalBayar(pembelian Pembelian) {
	if pembelian.TotalBayar > 0 {
		fmt.Printf("\nTerima kasih %s atas pembelian Anda!\n", pembelian.Nama)
		fmt.Printf("Barang: %s\n", pembelian.Barang)
		fmt.Printf("Jumlah: %d\n", pembelian.JumlahBeli)
		fmt.Printf("Total Bayar: %s\n", formatHarga(pembelian.TotalBayar))
	} else {
		fmt.Println("\nPembelian gagal. Cek kembali input Anda.")
	}
}

func postPembelian(pembelian Pembelian) error {
	url := apiURL + "/pembelian"
	payload, err := json.Marshal(pembelian)
	if err != nil {
		return fmt.Errorf("gagal mengubah data pembelian ke JSON: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("gagal membuat request POST: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("gagal mengirim request POST: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("gagal menyimpan pembelian: %s", body)
	}

	return nil
}

func updateStok(barangName string, jumlahBeli int) error {
	barang, err := getAllBarang()
	if err != nil {
		return fmt.Errorf("gagal mengambil data barang: %w", err)
	}

	var barangId int
	for _, b := range barang {
		if strings.ToLower(b.NamaBarang) == strings.ToLower(barangName) {
			barangId = b.IdBarang
			break
		}
	}

	if barangId == 0 {
		return fmt.Errorf("barang tidak ditemukan")
	}

	url := fmt.Sprintf("%s/updateStok/%d", apiURL, barangId)
	payload := strconv.Itoa(jumlahBeli)

	req, err := http.NewRequest("PUT", url, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("gagal membuat request PUT: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("gagal mengirim request PUT: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("gagal memperbarui stok: %s", body)
	}

	return nil
}
func getRiwayatPembeli(namaPembeli string) ([]Pembelian, error) {
	url := fmt.Sprintf("%s/riwayatPembeli/%s", apiURL, namaPembeli)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request GET: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengirim request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("gagal mendapatkan riwayat pembelian: %s", body)
	}

	var riwayat []Pembelian
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons body: %w", err)
	}

	err = json.Unmarshal(body, &riwayat)
	if err != nil {
		return nil, fmt.Errorf("gagal meng-unmarshal JSON: %w", err)
	}

	return riwayat, nil
}
