!include "MUI2.nsh"

; General
Name "StockTracker"
OutFile "StockTracker-1.0.0-windows-amd64-installer.exe"
InstallDir "$LOCALAPPDATA\StockTracker"
InstallDirRegKey HKCU "Software\StockTracker" "InstallDir"
RequestExecutionLevel user

; UI
!define MUI_ABORTWARNING

; Pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; Languages
!insertmacro MUI_LANGUAGE "SimpChinese"

; Installer Section
Section "StockTracker" SecMain
  SetOutPath "$INSTDIR"
  
  ; Main executable
  File "..\bin\StockTracker.exe"
  
  ; Create uninstaller
  WriteUninstaller "$INSTDIR\uninstall.exe"
  
  ; Registry entries
  WriteRegStr HKCU "Software\StockTracker" "InstallDir" "$INSTDIR"
  WriteRegStr HKCU "Software\StockTracker" "Version" "1.0.0"
  
  ; Uninstall registry
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\StockTracker" \
    "DisplayName" "StockTracker"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\StockTracker" \
    "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\StockTracker" \
    "DisplayVersion" "1.0.0"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\StockTracker" \
    "Publisher" "StockTracker"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\StockTracker" \
    "InstallLocation" "$INSTDIR"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\StockTracker" \
    "DisplayIcon" "$INSTDIR\StockTracker.exe"
  
  ; Start Menu shortcut
  CreateDirectory "$SMPROGRAMS\StockTracker"
  CreateShortcut "$SMPROGRAMS\StockTracker\StockTracker.lnk" "$INSTDIR\StockTracker.exe"
  CreateShortcut "$SMPROGRAMS\StockTracker\Uninstall.lnk" "$INSTDIR\uninstall.exe"
  
  ; Desktop shortcut
  CreateShortcut "$DESKTOP\StockTracker.lnk" "$INSTDIR\StockTracker.exe"
SectionEnd

; Uninstaller Section
Section "Uninstall"
  ; Remove files
  Delete "$INSTDIR\StockTracker.exe"
  Delete "$INSTDIR\uninstall.exe"
  
  ; Remove shortcuts
  Delete "$SMPROGRAMS\StockTracker\StockTracker.lnk"
  Delete "$SMPROGRAMS\StockTracker\Uninstall.lnk"
  Delete "$DESKTOP\StockTracker.lnk"
  RMDir "$SMPROGRAMS\StockTracker"
  
  ; Remove registry
  DeleteRegKey HKCU "Software\StockTracker"
  DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\StockTracker"
  
  ; Remove install directory
  RMDir "$INSTDIR"
SectionEnd
