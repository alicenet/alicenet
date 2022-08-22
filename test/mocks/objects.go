package mocks

import (
	objs "github.com/alicenet/alicenet/consensus/objs"
	"github.com/ethereum/go-ethereum/core/types"
)

// This file contains functions to generate commonly used object types you may use in tests
// All returned objects are valid objects sampled from dev environments

func NewMockSnapshotTx() *types.Transaction {
	tx := &types.Transaction{}
	err := tx.UnmarshalBinary([]byte{249, 2, 104, 10, 1, 131, 18, 1, 43, 148, 147, 233, 22, 154, 169, 143, 9, 29, 213, 8, 199, 72, 112, 129, 209, 175, 218, 183, 207, 69, 128, 185, 2, 4, 8, 202, 31, 37, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 192, 31, 127, 172, 22, 115, 42, 90, 205, 18, 11, 164, 170, 60, 91, 40, 65, 170, 211, 47, 172, 44, 247, 93, 58, 91, 198, 254, 95, 56, 0, 12, 145, 20, 223, 81, 247, 203, 176, 58, 150, 81, 152, 211, 30, 159, 117, 124, 215, 247, 243, 135, 30, 155, 76, 194, 237, 96, 57, 174, 171, 197, 56, 239, 144, 25, 233, 194, 32, 143, 161, 28, 164, 130, 145, 219, 25, 218, 129, 134, 165, 202, 50, 142, 130, 94, 240, 142, 111, 239, 190, 137, 174, 189, 13, 194, 74, 42, 34, 151, 20, 115, 75, 95, 78, 250, 216, 5, 12, 36, 204, 133, 118, 173, 38, 28, 11, 16, 64, 11, 204, 37, 233, 110, 62, 217, 0, 185, 87, 10, 8, 217, 193, 120, 71, 234, 204, 35, 244, 161, 62, 141, 172, 161, 81, 67, 113, 31, 239, 175, 43, 214, 190, 159, 190, 154, 114, 244, 103, 242, 81, 16, 72, 155, 182, 110, 97, 224, 118, 12, 104, 176, 249, 75, 65, 31, 38, 121, 84, 250, 182, 130, 6, 173, 138, 93, 151, 128, 66, 183, 244, 26, 236, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 176, 0, 0, 0, 0, 1, 0, 4, 0, 42, 0, 0, 0, 128, 0, 0, 0, 13, 0, 0, 0, 2, 1, 0, 0, 25, 0, 0, 0, 2, 1, 0, 0, 37, 0, 0, 0, 2, 1, 0, 0, 49, 0, 0, 0, 2, 1, 0, 0, 50, 171, 192, 119, 134, 26, 235, 84, 22, 163, 221, 169, 33, 30, 34, 238, 106, 134, 63, 174, 159, 196, 27, 93, 40, 240, 213, 78, 63, 60, 136, 19, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 146, 165, 234, 66, 41, 57, 251, 39, 205, 79, 12, 192, 104, 85, 222, 140, 4, 182, 251, 25, 98, 207, 142, 4, 151, 124, 26, 236, 107, 65, 132, 144, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 130, 10, 150, 160, 84, 134, 203, 96, 109, 15, 126, 59, 67, 91, 125, 35, 137, 183, 228, 77, 159, 89, 55, 147, 48, 169, 243, 241, 107, 106, 129, 124, 67, 39, 3, 87, 160, 26, 198, 90, 124, 79, 12, 41, 32, 208, 231, 229, 121, 191, 100, 34, 83, 138, 164, 161, 95, 40, 179, 148, 115, 146, 42, 203, 179, 3, 254, 170, 120})
	if err != nil {
		panic(err)
	}
	return tx
}

func NewMockSnapshotTx2() *types.Transaction {
	tx := &types.Transaction{}
	err := tx.UnmarshalBinary([]byte{249, 2, 104, 71, 1, 131, 17, 253, 211, 148, 147, 233, 22, 154, 169, 143, 9, 29, 213, 8, 199, 72, 112, 129, 209, 175, 218, 183, 207, 69, 128, 185, 2, 4, 8, 202, 31, 37, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 192, 31, 127, 172, 22, 115, 42, 90, 205, 18, 11, 164, 170, 60, 91, 40, 65, 170, 211, 47, 172, 44, 247, 93, 58, 91, 198, 254, 95, 56, 0, 12, 145, 20, 223, 81, 247, 203, 176, 58, 150, 81, 152, 211, 30, 159, 117, 124, 215, 247, 243, 135, 30, 155, 76, 194, 237, 96, 57, 174, 171, 197, 56, 239, 144, 25, 233, 194, 32, 143, 161, 28, 164, 130, 145, 219, 25, 218, 129, 134, 165, 202, 50, 142, 130, 94, 240, 142, 111, 239, 190, 137, 174, 189, 13, 194, 74, 42, 34, 151, 20, 115, 75, 95, 78, 250, 216, 5, 12, 36, 204, 133, 118, 173, 38, 28, 11, 16, 64, 11, 204, 37, 233, 110, 62, 217, 0, 185, 87, 33, 20, 84, 85, 187, 207, 26, 148, 12, 13, 189, 182, 4, 55, 196, 117, 11, 95, 140, 113, 54, 174, 179, 56, 255, 218, 5, 70, 213, 245, 212, 220, 32, 198, 248, 32, 74, 49, 22, 160, 119, 195, 77, 52, 2, 184, 3, 26, 31, 75, 209, 95, 96, 240, 235, 222, 247, 187, 151, 97, 108, 9, 193, 238, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 176, 0, 0, 0, 0, 1, 0, 4, 0, 42, 0, 0, 0, 64, 41, 0, 0, 13, 0, 0, 0, 2, 1, 0, 0, 25, 0, 0, 0, 2, 1, 0, 0, 37, 0, 0, 0, 2, 1, 0, 0, 49, 0, 0, 0, 2, 1, 0, 0, 133, 110, 44, 114, 195, 88, 111, 157, 230, 125, 12, 170, 108, 152, 221, 134, 131, 53, 161, 24, 144, 152, 133, 68, 2, 136, 174, 32, 200, 155, 208, 3, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 33, 206, 139, 107, 210, 85, 65, 155, 62, 47, 125, 155, 219, 147, 214, 116, 140, 60, 215, 13, 205, 72, 126, 44, 162, 135, 242, 205, 223, 70, 43, 86, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 130, 10, 150, 160, 12, 211, 101, 157, 67, 46, 6, 43, 145, 29, 4, 90, 222, 103, 153, 141, 255, 243, 107, 170, 50, 115, 107, 161, 254, 141, 92, 0, 111, 173, 14, 115, 160, 23, 9, 150, 162, 213, 211, 127, 105, 173, 65, 220, 117, 165, 167, 161, 51, 28, 157, 141, 190, 66, 74, 216, 195, 72, 63, 123, 210, 44, 24, 139, 101})
	if err != nil {
		panic(err)
	}
	return tx
}

func NewMockBlockHeader() *objs.BlockHeader {
	blockHeader := &objs.BlockHeader{}
	err := blockHeader.UnmarshalBinary([]byte{0, 0, 0, 0, 0, 0, 3, 0, 8, 0, 0, 0, 1, 0, 4, 0, 89, 0, 0, 0, 2, 6, 0, 0, 181, 0, 0, 0, 2, 0, 0, 0, 42, 0, 0, 0, 8, 3, 0, 0, 13, 0, 0, 0, 2, 1, 0, 0, 25, 0, 0, 0, 2, 1, 0, 0, 37, 0, 0, 0, 2, 1, 0, 0, 49, 0, 0, 0, 2, 1, 0, 0, 125, 56, 56, 255, 62, 64, 136, 59, 115, 108, 129, 228, 18, 133, 160, 220, 127, 56, 179, 7, 55, 215, 39, 111, 187, 195, 120, 118, 22, 203, 242, 201, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 206, 149, 115, 249, 102, 3, 185, 58, 36, 197, 107, 9, 40, 75, 179, 75, 161, 110, 58, 11, 133, 64, 12, 26, 144, 171, 90, 190, 173, 128, 60, 206, 31, 127, 172, 22, 115, 42, 90, 205, 18, 11, 164, 170, 60, 91, 40, 65, 170, 211, 47, 172, 44, 247, 93, 58, 91, 198, 254, 95, 56, 0, 12, 145, 20, 223, 81, 247, 203, 176, 58, 150, 81, 152, 211, 30, 159, 117, 124, 215, 247, 243, 135, 30, 155, 76, 194, 237, 96, 57, 174, 171, 197, 56, 239, 144, 25, 233, 194, 32, 143, 161, 28, 164, 130, 145, 219, 25, 218, 129, 134, 165, 202, 50, 142, 130, 94, 240, 142, 111, 239, 190, 137, 174, 189, 13, 194, 74, 42, 34, 151, 20, 115, 75, 95, 78, 250, 216, 5, 12, 36, 204, 133, 118, 173, 38, 28, 11, 16, 64, 11, 204, 37, 233, 110, 62, 217, 0, 185, 87, 16, 199, 104, 254, 97, 159, 166, 106, 239, 71, 255, 158, 63, 80, 171, 211, 175, 43, 93, 114, 134, 3, 0, 211, 177, 136, 111, 26, 22, 108, 118, 199, 29, 106, 82, 61, 246, 187, 142, 226, 14, 167, 116, 171, 194, 244, 157, 203, 217, 127, 150, 130, 49, 129, 224, 242, 60, 229, 35, 70, 107, 245, 20, 122})
	if err != nil {
		panic(err)
	}
	return blockHeader
}

func NewMockOwnState() *objs.OwnState {
	os := &objs.OwnState{}
	err := os.UnmarshalBinary([]byte{0, 0, 0, 0, 0, 0, 6, 0, 21, 0, 0, 0, 162, 0, 0, 0, 29, 0, 0, 0, 2, 4, 0, 0, 88, 0, 0, 0, 0, 0, 3, 0, 20, 1, 0, 0, 0, 0, 3, 0, 208, 1, 0, 0, 0, 0, 3, 0, 140, 2, 0, 0, 0, 0, 3, 0, 99, 166, 98, 123, 121, 129, 58, 122, 67, 130, 148, 144, 196, 206, 64, 146, 84, 246, 65, 119, 0, 0, 0, 0, 31, 127, 172, 22, 115, 42, 90, 205, 18, 11, 164, 170, 60, 91, 40, 65, 170, 211, 47, 172, 44, 247, 93, 58, 91, 198, 254, 95, 56, 0, 12, 145, 20, 223, 81, 247, 203, 176, 58, 150, 81, 152, 211, 30, 159, 117, 124, 215, 247, 243, 135, 30, 155, 76, 194, 237, 96, 57, 174, 171, 197, 56, 239, 144, 25, 233, 194, 32, 143, 161, 28, 164, 130, 145, 219, 25, 218, 129, 134, 165, 202, 50, 142, 130, 94, 240, 142, 111, 239, 190, 137, 174, 189, 13, 194, 74, 42, 34, 151, 20, 115, 75, 95, 78, 250, 216, 5, 12, 36, 204, 133, 118, 173, 38, 28, 11, 16, 64, 11, 204, 37, 233, 110, 62, 217, 0, 185, 87, 8, 0, 0, 0, 1, 0, 4, 0, 89, 0, 0, 0, 2, 6, 0, 0, 181, 0, 0, 0, 2, 0, 0, 0, 42, 0, 0, 0, 57, 64, 0, 0, 13, 0, 0, 0, 2, 1, 0, 0, 25, 0, 0, 0, 2, 1, 0, 0, 37, 0, 0, 0, 2, 1, 0, 0, 49, 0, 0, 0, 2, 1, 0, 0, 234, 242, 97, 234, 15, 234, 155, 1, 248, 63, 59, 148, 56, 53, 49, 60, 201, 31, 11, 123, 61, 225, 139, 194, 122, 95, 250, 156, 40, 163, 107, 162, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 12, 33, 167, 224, 57, 253, 212, 29, 138, 163, 127, 176, 200, 37, 169, 188, 67, 58, 169, 65, 234, 252, 173, 175, 145, 34, 211, 178, 142, 229, 12, 2, 31, 127, 172, 22, 115, 42, 90, 205, 18, 11, 164, 170, 60, 91, 40, 65, 170, 211, 47, 172, 44, 247, 93, 58, 91, 198, 254, 95, 56, 0, 12, 145, 20, 223, 81, 247, 203, 176, 58, 150, 81, 152, 211, 30, 159, 117, 124, 215, 247, 243, 135, 30, 155, 76, 194, 237, 96, 57, 174, 171, 197, 56, 239, 144, 25, 233, 194, 32, 143, 161, 28, 164, 130, 145, 219, 25, 218, 129, 134, 165, 202, 50, 142, 130, 94, 240, 142, 111, 239, 190, 137, 174, 189, 13, 194, 74, 42, 34, 151, 20, 115, 75, 95, 78, 250, 216, 5, 12, 36, 204, 133, 118, 173, 38, 28, 11, 16, 64, 11, 204, 37, 233, 110, 62, 217, 0, 185, 87, 18, 121, 127, 138, 231, 248, 220, 99, 125, 126, 250, 94, 215, 165, 137, 165, 177, 123, 60, 29, 121, 241, 21, 21, 182, 228, 62, 252, 91, 161, 152, 170, 36, 189, 82, 127, 183, 119, 174, 151, 56, 170, 74, 49, 112, 115, 19, 48, 166, 99, 177, 53, 30, 130, 152, 114, 247, 141, 212, 44, 63, 45, 36, 71, 8, 0, 0, 0, 1, 0, 4, 0, 89, 0, 0, 0, 2, 6, 0, 0, 181, 0, 0, 0, 2, 0, 0, 0, 42, 0, 0, 0, 57, 64, 0, 0, 13, 0, 0, 0, 2, 1, 0, 0, 25, 0, 0, 0, 2, 1, 0, 0, 37, 0, 0, 0, 2, 1, 0, 0, 49, 0, 0, 0, 2, 1, 0, 0, 234, 242, 97, 234, 15, 234, 155, 1, 248, 63, 59, 148, 56, 53, 49, 60, 201, 31, 11, 123, 61, 225, 139, 194, 122, 95, 250, 156, 40, 163, 107, 162, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 12, 33, 167, 224, 57, 253, 212, 29, 138, 163, 127, 176, 200, 37, 169, 188, 67, 58, 169, 65, 234, 252, 173, 175, 145, 34, 211, 178, 142, 229, 12, 2, 31, 127, 172, 22, 115, 42, 90, 205, 18, 11, 164, 170, 60, 91, 40, 65, 170, 211, 47, 172, 44, 247, 93, 58, 91, 198, 254, 95, 56, 0, 12, 145, 20, 223, 81, 247, 203, 176, 58, 150, 81, 152, 211, 30, 159, 117, 124, 215, 247, 243, 135, 30, 155, 76, 194, 237, 96, 57, 174, 171, 197, 56, 239, 144, 25, 233, 194, 32, 143, 161, 28, 164, 130, 145, 219, 25, 218, 129, 134, 165, 202, 50, 142, 130, 94, 240, 142, 111, 239, 190, 137, 174, 189, 13, 194, 74, 42, 34, 151, 20, 115, 75, 95, 78, 250, 216, 5, 12, 36, 204, 133, 118, 173, 38, 28, 11, 16, 64, 11, 204, 37, 233, 110, 62, 217, 0, 185, 87, 18, 121, 127, 138, 231, 248, 220, 99, 125, 126, 250, 94, 215, 165, 137, 165, 177, 123, 60, 29, 121, 241, 21, 21, 182, 228, 62, 252, 91, 161, 152, 170, 36, 189, 82, 127, 183, 119, 174, 151, 56, 170, 74, 49, 112, 115, 19, 48, 166, 99, 177, 53, 30, 130, 152, 114, 247, 141, 212, 44, 63, 45, 36, 71, 8, 0, 0, 0, 1, 0, 4, 0, 89, 0, 0, 0, 2, 6, 0, 0, 181, 0, 0, 0, 2, 0, 0, 0, 42, 0, 0, 0, 0, 64, 0, 0, 13, 0, 0, 0, 2, 1, 0, 0, 25, 0, 0, 0, 2, 1, 0, 0, 37, 0, 0, 0, 2, 1, 0, 0, 49, 0, 0, 0, 2, 1, 0, 0, 244, 95, 247, 75, 101, 23, 15, 161, 22, 93, 46, 6, 241, 172, 187, 13, 54, 175, 70, 201, 14, 0, 216, 55, 19, 234, 138, 15, 110, 61, 152, 228, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 26, 9, 230, 149, 32, 235, 199, 30, 56, 217, 189, 218, 168, 54, 129, 68, 158, 79, 172, 52, 84, 216, 169, 196, 125, 154, 108, 132, 98, 103, 39, 59, 31, 127, 172, 22, 115, 42, 90, 205, 18, 11, 164, 170, 60, 91, 40, 65, 170, 211, 47, 172, 44, 247, 93, 58, 91, 198, 254, 95, 56, 0, 12, 145, 20, 223, 81, 247, 203, 176, 58, 150, 81, 152, 211, 30, 159, 117, 124, 215, 247, 243, 135, 30, 155, 76, 194, 237, 96, 57, 174, 171, 197, 56, 239, 144, 25, 233, 194, 32, 143, 161, 28, 164, 130, 145, 219, 25, 218, 129, 134, 165, 202, 50, 142, 130, 94, 240, 142, 111, 239, 190, 137, 174, 189, 13, 194, 74, 42, 34, 151, 20, 115, 75, 95, 78, 250, 216, 5, 12, 36, 204, 133, 118, 173, 38, 28, 11, 16, 64, 11, 204, 37, 233, 110, 62, 217, 0, 185, 87, 0, 205, 51, 109, 23, 86, 65, 83, 238, 176, 252, 107, 182, 79, 19, 197, 240, 43, 255, 16, 9, 156, 21, 20, 151, 148, 35, 32, 99, 6, 173, 84, 21, 55, 206, 76, 131, 247, 20, 58, 173, 157, 65, 143, 138, 194, 226, 217, 133, 46, 163, 115, 152, 191, 144, 9, 189, 202, 186, 58, 209, 45, 148, 11, 8, 0, 0, 0, 1, 0, 4, 0, 89, 0, 0, 0, 2, 6, 0, 0, 181, 0, 0, 0, 2, 0, 0, 0, 42, 0, 0, 0, 32, 64, 0, 0, 13, 0, 0, 0, 2, 1, 0, 0, 25, 0, 0, 0, 2, 1, 0, 0, 37, 0, 0, 0, 2, 1, 0, 0, 49, 0, 0, 0, 2, 1, 0, 0, 150, 113, 88, 97, 227, 17, 205, 23, 6, 64, 193, 100, 140, 101, 47, 186, 241, 211, 179, 249, 150, 192, 8, 101, 218, 114, 34, 30, 46, 242, 75, 232, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 21, 169, 208, 212, 249, 74, 229, 167, 138, 83, 235, 141, 145, 222, 112, 72, 114, 175, 159, 250, 248, 10, 151, 77, 121, 52, 100, 95, 133, 86, 58, 91, 31, 127, 172, 22, 115, 42, 90, 205, 18, 11, 164, 170, 60, 91, 40, 65, 170, 211, 47, 172, 44, 247, 93, 58, 91, 198, 254, 95, 56, 0, 12, 145, 20, 223, 81, 247, 203, 176, 58, 150, 81, 152, 211, 30, 159, 117, 124, 215, 247, 243, 135, 30, 155, 76, 194, 237, 96, 57, 174, 171, 197, 56, 239, 144, 25, 233, 194, 32, 143, 161, 28, 164, 130, 145, 219, 25, 218, 129, 134, 165, 202, 50, 142, 130, 94, 240, 142, 111, 239, 190, 137, 174, 189, 13, 194, 74, 42, 34, 151, 20, 115, 75, 95, 78, 250, 216, 5, 12, 36, 204, 133, 118, 173, 38, 28, 11, 16, 64, 11, 204, 37, 233, 110, 62, 217, 0, 185, 87, 10, 246, 236, 34, 98, 56, 84, 55, 216, 94, 106, 19, 128, 157, 153, 251, 108, 57, 77, 182, 152, 11, 21, 137, 170, 53, 164, 210, 76, 74, 21, 72, 9, 93, 29, 212, 155, 62, 229, 58, 57, 155, 171, 218, 71, 204, 27, 159, 105, 146, 251, 160, 164, 32, 21, 214, 68, 169, 193, 185, 121, 221, 126, 83})
	if err != nil {
		panic(err)
	}
	return os
}
