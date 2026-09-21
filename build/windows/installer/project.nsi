Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them with the values from ProjectInfo.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows to use this project.nsi manually
## from outside of Wails for debugging and development of the installer.
##
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the ProjectInfo file, but they can be overwritten here.
####
## !define INFO_PROJECTNAME    "MyProject" # Default "{{.Name}}"
## !define INFO_COMPANYNAME    "MyCompany" # Default "{{.Info.CompanyName}}"
## !define INFO_PRODUCTNAME    "MyProduct" # Default "{{.Info.ProductName}}"
## !define INFO_PRODUCTVERSION "1.0.0"     # Default "{{.Info.ProductVersion}}"
## !define INFO_COPYRIGHT      "Copyright" # Default "{{.Info.Copyright}}"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
## !define REQUEST_EXECUTION_LEVEL "admin"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
####
## Include the wails tools
####
!include "wails_tools.nsh"

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.

; Embedded helper used for read-only machine validation before installation.
!ifndef MAJUCAU_HELPER_SOURCE
  !define MAJUCAU_HELPER_SOURCE "..\..\bin\installer-helper.exe"
!endif
!ifndef MAJUCAU_DIAGNOSTICS_DIR
  !define MAJUCAU_DIAGNOSTICS_DIR "$TEMP\Majucau"
!endif
!ifndef MAJUCAU_DATA_DIR
  !define MAJUCAU_DATA_DIR "$PROGRAMDATA\Majucau"
!endif
!ifndef MAJUCAU_PREFERRED_PORT
  !define MAJUCAU_PREFERRED_PORT "54329"
!endif
!ifndef MAJUCAU_MANIFEST_SOURCE
  !define MAJUCAU_MANIFEST_SOURCE "..\..\..\artifacts\windows\release-manifest.json"
!endif
!ifndef MAJUCAU_MIGRATION_SOURCE
  !define MAJUCAU_MIGRATION_SOURCE "..\..\..\artifacts\windows\migrations"
!endif
!ifndef MAJUCAU_DESKTOP_SOURCE
  !define MAJUCAU_DESKTOP_SOURCE "..\..\bin\majucau.exe"
!endif
!ifndef MAJUCAU_WORKER_SOURCE
  !define MAJUCAU_WORKER_SOURCE "..\..\bin\majucau-worker.exe"
!endif

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # Adds a EULA page to the installer
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uinstalling page

!insertmacro MUI_LANGUAGE "PortugueseBR"

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
!ifdef WAILS_INSTALL_SCOPE
  !if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
  !else
    InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
  !endif
!else
  InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!endif # Default installing folder ($PROGRAMFILES is Program Files folder).
ShowInstDetails show # This will always show the installation details.

Function .onInit
   !insertmacro wails.checkArchitecture

   ; Run the same read-only checks used by the portable smoke test. A failed
   ; preflight stops installation and leaves a sanitized diagnostic bundle.
   InitPluginsDir
   File /oname=$PLUGINSDIR\majucau-installer-helper.exe "${MAJUCAU_HELPER_SOURCE}"
   CreateDirectory "${MAJUCAU_DIAGNOSTICS_DIR}"
   StrCpy $0 "${MAJUCAU_DIAGNOSTICS_DIR}\installer-preflight.zip"
   ExecWait '"$PLUGINSDIR\majucau-installer-helper.exe" preflight --install-dir "$INSTDIR" --data-dir "${MAJUCAU_DATA_DIR}" --free-space-path "$PROGRAMDATA" --port ${MAJUCAU_PREFERRED_PORT}' $1
   ${If} $1 != 0
       ExecWait '"$PLUGINSDIR\majucau-installer-helper.exe" diagnostics --output "$0" --install-dir "$INSTDIR" --data-dir "${MAJUCAU_DATA_DIR}" --free-space-path "$PROGRAMDATA" --port ${MAJUCAU_PREFERRED_PORT}' $2
       IfSilent MajuauPreflightSilent MajuauPreflightInteractive
       MajuauPreflightSilent:
           SetErrorLevel $1
           Quit
       MajuauPreflightInteractive:
       ${If} $1 == 2
           MessageBox MB_ICONSTOP|MB_OK "A instalação não pode continuar porque a máquina não passou no preflight.\n\nO diagnóstico sanitizado foi salvo em:\n$0\n\nCorrija os itens indicados e execute o instalador novamente."
       ${Else}
           MessageBox MB_ICONSTOP|MB_OK "Não foi possível validar a máquina para a instalação (código $1).\n\nO diagnóstico sanitizado foi salvo em:\n$0"
       ${EndIf}
       Quit
   ${EndIf}

   ; Verify the exact embedded package before any installation side effect.
   ; The manifest contains only hashes, schema compatibility and migration
   ; inventory; it never contains credentials or user data.
   CreateDirectory "$PLUGINSDIR\package"
   CreateDirectory "$PLUGINSDIR\package\migrations"
   File /oname=$PLUGINSDIR\package\release-manifest.json "${MAJUCAU_MANIFEST_SOURCE}"
   File /oname=$PLUGINSDIR\package\majucau.exe "${MAJUCAU_DESKTOP_SOURCE}"
   File /oname=$PLUGINSDIR\package\majucau-worker.exe "${MAJUCAU_WORKER_SOURCE}"
   File /oname=$PLUGINSDIR\package\installer-helper.exe "${MAJUCAU_HELPER_SOURCE}"
   File /oname=$PLUGINSDIR\package\migrations\000001_init.up.sql "${MAJUCAU_MIGRATION_SOURCE}\000001_init.up.sql"
   ExecWait '"$PLUGINSDIR\majucau-installer-helper.exe" verify-package --package-dir "$PLUGINSDIR\package" --manifest "release-manifest.json" --current-schema "1.0"' $3
   ${If} $3 != 0
       IfSilent MajuauPackageSilent MajuauPackageInteractive
       MajuauPackageSilent:
           SetErrorLevel $3
           Quit
       MajuauPackageInteractive:
           MessageBox MB_ICONSTOP|MB_OK "O pacote do instalador não passou na verificação de integridade (código $3).\n\nO instalador foi bloqueado para proteger a instalação."
           Quit
   ${EndIf}
FunctionEnd

Section
    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files
    File "/oname=majucau-worker.exe" "${MAJUCAU_WORKER_SOURCE}"
    File "/oname=installer-helper.exe" "${MAJUCAU_HELPER_SOURCE}"

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd
