package main

import (
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"

	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/pierrec/lz4"
)

var sdkPath string
var gamePath string
var vtexPath string
var imgPath string
var vpcfPath string

var vtexCRC32 uint32

func main() {

	// Get terminal arguments.
	flag.StringVar(&sdkPath, "sdk", "", "Path to the top directory in your SDK installation. Defaults to the directory the executable is in.")
	flag.StringVar(&gamePath, "game", "", "The game directory in your SDK installation to use. Defaults to \"hlvr\"")
	flag.StringVar(&vtexPath, "vtex", "", "game directory relative path for the .vtex file.")
	flag.StringVar(&imgPath, "img", "", "Path to image file to convert to a .vtex file.")
	flag.StringVar(&vpcfPath, "vpcf", "", "game directory relative path to output the .vpcf file.")

	flag.Parse()

	if sdkPath == "" {
		execPath, err := os.Executable()
		sdkPath = filepath.Dir(execPath)
		if err != nil {
			fmt.Printf("Error: Unable to get SDK location. Reason: %s\n", err)
			os.Exit(2)
		}
	}
	sdkPath = filepath.Join(sdkPath, "game")

	if gamePath == "" {
		fmt.Printf("Warning: Not specified game path. Using \"hlvr\".\n")
		gamePath = "hlvr"
	}

	if vtexPath == "" {
		fmt.Printf("Error: You need to specify a .vtex path using the -vtex argument.\n")
		os.Exit(2)
	}

	vtexPathAbs := filepath.Join(sdkPath, gamePath, vtexPath+"_c")
	vtexBytes, err := ioutil.ReadFile(vtexPathAbs)
	if err != nil {
		fmt.Printf("Error: Failed to read .vtex file. Reason: %s\n", vtexPathAbs, err)
		os.Exit(2)
	}

	fmt.Printf("Info: Read .vtex file \"%s\"\n", vtexPathAbs)
	vtexCRC32 = crc32.ChecksumIEEE(vtexBytes)
	fmt.Printf("Info: CRC-32 checksum = 0x%08x.\n", vtexCRC32)

	if vpcfPath == "" {
		filename := filepath.Base(vtexPath)
		filename = strings.TrimSuffix(filename, filepath.Ext(filename))
		vpcfPath = filepath.Join("particles/dry_erase/", filename+".vpcf")
	}

	bytes, err := createVpcfFile()
	if err != nil {
		fmt.Printf("Error: Failed to create .vpfc file. Reason: %s\n", err)
		os.Exit(2)
	}

	// Setup  output path.
	vpcfPathAbs := filepath.Join(sdkPath, gamePath, vpcfPath+"_c")
	err = os.MkdirAll(filepath.Dir(vpcfPathAbs), os.ModePerm)
	if err != nil {
		fmt.Printf("Error: Failed to save .vpcf file. Reason: %s\n", err)
	}

	// Write .vpcf file.
	err = ioutil.WriteFile(vpcfPathAbs, bytes, 0644)
	if err != nil {
		fmt.Printf("Error: Failed to write .vpcf file. Reason: %s\n", err)
		os.Exit(2)
	}

	fmt.Printf("Info: Created new .vpcf file \"%s\"\n", vpcfPathAbs)
}

//func createVtexFile() ([]byte, error) {
//
//	// Read image file bytes.
//	imgData, err := ioutil.ReadFile(imgPath)
//
//	Use(imgData)
//
//	if err != nil {
//		return nil, fmt.Errorf("Failed to read image file (%s). Reason: %s.", imgPath, err)
//	}
//
//	return nil, nil
//}

func createVpcfFile() ([]byte, error) {

	// Setup
	vpcfHeader, _ := base64.StdEncoding.DecodeString(VPCF_HEADER)
	vpcfData, _ := base64.StdEncoding.DecodeString(VPCF_DATA)

	vpcfPathRaw := []byte(filepath.ToSlash(vpcfPath))
	vtexPathRaw := []byte(filepath.ToSlash(vtexPath))
	gamePathRaw := []byte(filepath.ToSlash(gamePath))

	vpcfPathLen := uint32(len(vpcfPathRaw))
	vtexPathLen := uint32(len(vtexPathRaw))
	gamePathLen := uint32(len(gamePathRaw))

	bytes := make([]byte, 0, 2500)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rriId := r.Uint64()

	// Create file
	bytes = append(bytes, vpcfHeader[:VPCF_RERL_VTEX_PATH_PTR]...)
	bytes = append(bytes, vtexPathRaw...)
	bytes = append(bytes, vpcfHeader[VPCF_RERL_VTEX_PATH_PTR:VPCF_REDI_CRF_PTR]...)
	bytes = append(bytes, vpcfPathRaw...)
	bytes = append(bytes, vpcfHeader[VPCF_REDI_CRF_PTR:VPCF_REDI_CSP_PTR]...)
	bytes = append(bytes, gamePathRaw...)
	bytes = append(bytes, vpcfHeader[VPCF_REDI_CSP_PTR:]...)

	binary.LittleEndian.PutUint32(bytes[VPCF_HEADER_RERL_SIZE_PTR:], vtexPathLen+0x1D)
	binary.LittleEndian.PutUint32(bytes[VPCF_HEADER_REDI_OFFSET_PTR:], vtexPathLen+0x31)
	binary.LittleEndian.PutUint64(bytes[VPCF_RERL_RRI_ID_PTR:], rriId)

	binary.LittleEndian.PutUint32(bytes[VPCF_RERL_VTEX_PATH_PTR+int(vtexPathLen)+0x11:], gamePathLen+vpcfPathLen+0x52)
	binary.LittleEndian.PutUint32(bytes[VPCF_RERL_VTEX_PATH_PTR+int(vtexPathLen)+0x19:], gamePathLen+vpcfPathLen+0x82)
	binary.LittleEndian.PutUint32(bytes[VPCF_RERL_VTEX_PATH_PTR+int(vtexPathLen)+0x39:], gamePathLen+vpcfPathLen+0x9e)

	binary.LittleEndian.PutUint32(bytes[VPCF_REDI_CSP_PTR_PTR+int(vtexPathLen):], vpcfPathLen+0x0d)
	binary.LittleEndian.PutUint32(bytes[VPCF_REDI_CRC32_PTR+int(vtexPathLen):], vtexCRC32)

	// Data section
	binary.LittleEndian.PutUint32(bytes[VPCF_HEADER_DATA_OFFSET_PTR:], uint32(len(bytes))-0x50)

	bytes = append(bytes, 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(bytes[len(bytes)-4:], uint32(len(vpcfData))+vtexPathLen)

	dataSrc := make([]byte, 0, len(vpcfData)+int(vtexPathLen))
	dataDst := make([]byte, cap(dataSrc))

	dataSrc = append(dataSrc, vpcfData[:VPCF_DATA_TEXTURE_PTR]...)
	dataSrc = append(dataSrc, vtexPathRaw...)
	dataSrc = append(dataSrc, vpcfData[VPCF_DATA_TEXTURE_PTR:]...)

	n, err := lz4.CompressBlock(dataSrc, dataDst, nil)

	if err != nil {
		return nil, err
	}

	bytes = append(bytes, dataDst[:n]...)

	// Update header again
	binary.LittleEndian.PutUint32(bytes[VPCF_HEADER_DATA_SIZE_PTR:], uint32(n+0x28))
	binary.LittleEndian.PutUint32(bytes[VPCF_HEADER_FILESIZE_PTR:], uint32(len(bytes)))

	return bytes, nil
}

func Use(vals ...interface{}) {
	for _, val := range vals {
		_ = val
	}
}
