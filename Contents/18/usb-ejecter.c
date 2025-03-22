#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <windows.h>
#include <Setupapi.h>
#include <winioctl.h>
#include <cfgmgr32.h>

#pragma comment(lib, "setupapi.lib")

#ifndef HEADER_H
#define HEADER_H

#define MAX_EJECT_TRIES 10
#define EJECT_TIMEOUT_MS 500

// GUIDs for device interfaces
const GUID GUID_DEVINTERFACE_VOLUME = {0x53f5630d, 0xb6bf, 0x11d0, {0x94, 0xf2, 0x00, 0xa0, 0xc9, 0x1e, 0xfb, 0x8b}};
const GUID GUID_DEVINTERFACE_DISK = {0x53f56307, 0xb6bf, 0x11d0, {0x94, 0xf2, 0x00, 0xa0, 0xc9, 0x1e, 0xfb, 0x8b}};
const GUID GUID_DEVINTERFACE_FLOPPY = {0x53f56311, 0xb6bf, 0x11d0, {0x94, 0xf2, 0x00, 0xa0, 0xc9, 0x1e, 0xfb, 0x8b}};
const GUID GUID_DEVINTERFACE_CDROM = {0x53f56308, 0xb6bf, 0x11d0, {0x94, 0xf2, 0x00, 0xa0, 0xc9, 0x1e, 0xfb, 0x8b}};

// Function prototypes
bool EjectDriveByLetter(wchar_t letter);
bool EjectDriveByLetterEx(wchar_t letter, DWORD tries, DWORD timeout);
DEVINST GetDrivesDevInst(long deviceNumber, UINT driveType, const wchar_t *dosDeviceName);
void PrintDriveType(UINT driveType);

#endif

//export EjectDriveByLetter
bool EjectDriveByLetter(wchar_t letter) {
	return EjectDriveByLetterEx(letter, MAX_EJECT_TRIES, EJECT_TIMEOUT_MS);
}

// Eject a drive by letter with custom parameters
bool EjectDriveByLetterEx(wchar_t letter, DWORD tries, DWORD timeout) {
	letter = towupper(letter); // Convert to uppercase

	if (letter < L'A' || letter > L'Z') {
		return false; // Invalid drive letter
	}

	wchar_t rootPath[] = L"X:\\";
	wchar_t devicePath[] = L"X:";
	wchar_t volumeAccessPath[] = L"\\\\.\\X:";

	rootPath[0] = letter;
	devicePath[0] = letter;
	volumeAccessPath[4] = letter;

	long deviceNumber = -1;
	HANDLE volumeHandle = CreateFileW(volumeAccessPath, 0, FILE_SHARE_READ | FILE_SHARE_WRITE, NULL, OPEN_EXISTING, 0, NULL);

	if (volumeHandle == INVALID_HANDLE_VALUE) {
		return false; // Failed to open volume
	}

	STORAGE_DEVICE_NUMBER storageDeviceNumber = {0};
	DWORD bytesReturned = 0;

	if (DeviceIoControl(volumeHandle, IOCTL_STORAGE_GET_DEVICE_NUMBER, NULL, 0, &storageDeviceNumber, sizeof(storageDeviceNumber), &bytesReturned, NULL)) {
		deviceNumber = storageDeviceNumber.DeviceNumber;
	}

	CloseHandle(volumeHandle);

	if (deviceNumber == -1) {
		return false; // Failed to get device number
	}

	UINT driveType = GetDriveTypeW(rootPath);
	wchar_t dosDeviceName[MAX_PATH];

	if (QueryDosDeviceW(devicePath, dosDeviceName, MAX_PATH) == 0) {
		return false; // Failed to query DOS device name
	}

	DEVINST deviceInstance = GetDrivesDevInst(deviceNumber, driveType, dosDeviceName);

	if (deviceInstance == 0) {
		return false; // Failed to get device instance
	}

	DEVINST parentDeviceInstance = 0;
	CM_Get_Parent(&parentDeviceInstance, deviceInstance, 0);

	bool ejectedSuccessfully = false;
	PNP_VETO_TYPE vetoType = PNP_VetoTypeUnknown;
	wchar_t vetoName[MAX_PATH] = {0};

	for (DWORD i = 0; i < tries; i++) {
		CONFIGRET configResult = CM_Request_Device_EjectW(parentDeviceInstance, &vetoType, vetoName, MAX_PATH, 0);
		ejectedSuccessfully = (configResult == CR_SUCCESS && vetoType == PNP_VetoTypeUnknown);

		if (ejectedSuccessfully) {
			break; // Ejection successful
		}

		Sleep(timeout); // Wait and retry
	}

	return ejectedSuccessfully;
}

// Get device instance handle
DEVINST GetDrivesDevInst(long deviceNumber, UINT driveType, const wchar_t *dosDeviceName) {
	bool isFloppy = (wcsstr(dosDeviceName, L"\\Floppy") != NULL);
	const GUID *deviceInterfaceGuid = NULL;

	switch (driveType) {
		case DRIVE_REMOVABLE:
			deviceInterfaceGuid = (isFloppy) ? &GUID_DEVINTERFACE_FLOPPY : &GUID_DEVINTERFACE_DISK;
			break;
		case DRIVE_FIXED:
			deviceInterfaceGuid = &GUID_DEVINTERFACE_DISK;
			break;
		case DRIVE_CDROM:
			deviceInterfaceGuid = &GUID_DEVINTERFACE_CDROM;
			break;
		default:
			return 0; // Invalid drive type
	}

	HDEVINFO deviceInfoSet = SetupDiGetClassDevsW(deviceInterfaceGuid, NULL, NULL, DIGCF_PRESENT | DIGCF_DEVICEINTERFACE);

	if (deviceInfoSet == INVALID_HANDLE_VALUE) {
		return 0; // Failed to get device information set
	}

	SP_DEVICE_INTERFACE_DATA deviceInterfaceData = {sizeof(deviceInterfaceData)};
	SP_DEVINFO_DATA deviceInfoData = {sizeof(deviceInfoData)};
	BYTE buffer[1024];
	PSP_DEVICE_INTERFACE_DETAIL_DATA_W deviceInterfaceDetailData = (PSP_DEVICE_INTERFACE_DETAIL_DATA_W)buffer;
	DWORD requiredSize = 0;

	for (DWORD index = 0; SetupDiEnumDeviceInterfaces(deviceInfoSet, NULL, deviceInterfaceGuid, index, &deviceInterfaceData); index++) {
		SetupDiGetDeviceInterfaceDetailW(deviceInfoSet, &deviceInterfaceData, NULL, 0, &requiredSize, NULL);

		if (requiredSize == 0 || requiredSize > sizeof(buffer)) {
			continue; // Skip if size check fails
		}

		deviceInterfaceDetailData->cbSize = sizeof(*deviceInterfaceDetailData);

		if (SetupDiGetDeviceInterfaceDetailW(deviceInfoSet, &deviceInterfaceData, deviceInterfaceDetailData, requiredSize, &requiredSize, &deviceInfoData)) {
			HANDLE driveHandle = CreateFileW(deviceInterfaceDetailData->DevicePath, 0, FILE_SHARE_READ | FILE_SHARE_WRITE, NULL, OPEN_EXISTING, 0, NULL);

			if (driveHandle != INVALID_HANDLE_VALUE) {
				STORAGE_DEVICE_NUMBER currentStorageDeviceNumber;
				DWORD bytesReturned = 0;

				if (DeviceIoControl(driveHandle, IOCTL_STORAGE_GET_DEVICE_NUMBER, NULL, 0, &currentStorageDeviceNumber, sizeof(currentStorageDeviceNumber), &bytesReturned, NULL)) {
					if (deviceNumber == (long)currentStorageDeviceNumber.DeviceNumber) {
						CloseHandle(driveHandle);
						SetupDiDestroyDeviceInfoList(deviceInfoSet);
						return deviceInfoData.DevInst; // Found the matching device
					}
				}
				CloseHandle(driveHandle);
			}
		}
	}

	SetupDiDestroyDeviceInfoList(deviceInfoSet);
	return 0; // Device not found
}

// Print drive type
void PrintDriveType(UINT driveType) {
	wprintf(L"Drive type is ");

	switch (driveType) {
		case DRIVE_UNKNOWN: wprintf(L"DRIVE_UNKNOWN"); break;
		case DRIVE_NO_ROOT_DIR: wprintf(L"DRIVE_NO_ROOT_DIR"); break;
		case DRIVE_REMOVABLE: wprintf(L"DRIVE_REMOVABLE"); break;
		case DRIVE_FIXED: wprintf(L"DRIVE_FIXED"); break;
		case DRIVE_REMOTE: wprintf(L"DRIVE_REMOTE"); break;
		case DRIVE_CDROM: wprintf(L"DRIVE_CDROM"); break;
		case DRIVE_RAMDISK: wprintf(L"DRIVE_RAMDISK"); break;
		default: wprintf(L"<Unknown drive type>"); break;
	}

	wprintf(L".\n");
}
