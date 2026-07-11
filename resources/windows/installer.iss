; JobTracker Windows installer.
; Built by CI (release.yml) with Inno Setup:
;   ISCC /DAppVersion=<x.y.z> /DAppArch=<amd64|arm64> /DBinary=<built-exe> installer.iss
; Paths are relative to this file (resources/windows/).

[Setup]
AppId=net.tecnologer.jobtracker
AppName=JobTracker
AppVersion={#AppVersion}
AppPublisher=Tecnologer
DefaultDirName={autopf}\JobTracker
DefaultGroupName=JobTracker
LicenseFile=..\..\LICENSE
OutputDir=..\..
OutputBaseFilename=jobtracker-desktop-windows-{#AppArch}-setup
#if AppArch == "arm64"
ArchitecturesAllowed=arm64
ArchitecturesInstallIn64BitMode=arm64
#else
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
#endif

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "..\..\{#Binary}"; DestDir: "{app}"; DestName: "jobtracker-desktop.exe"; Flags: ignoreversion
Source: "..\..\LICENSE"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\JobTracker"; Filename: "{app}\jobtracker-desktop.exe"
Name: "{autodesktop}\JobTracker"; Filename: "{app}\jobtracker-desktop.exe"; Tasks: desktopicon
