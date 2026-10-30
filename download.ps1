$medienOrdner = "Medien"
if (-not (Test-Path $medienOrdner)) { New-Item -ItemType Directory -Path $medienOrdner | Out-Null }

$datei = "output\output_app_bsky_feed_post.json"
$userDid = "did:plc:m52sdcq435oevizuq5rwqa6l"
# Deine Instanz-Domain:
$pdsHost = "eurosky.social" 

if (-not (Test-Path $datei)) {
    Write-Host "FEHLER: Datei $datei wurde nicht gefunden!" -ForegroundColor Red
    exit
}

Write-Host "Lese JSON-Datei ein..." -ForegroundColor Cyan
$posts = Get-Content $datei -Encoding utf8 -Raw | ConvertFrom-Json

$medien = foreach ($p in $posts) {
    if ($p.data.embed -and $p.data.embed.images) {
        
        $datumRaw = $p.data.createdAt
        $datumClean = "Unbekannt"
        if ($datumRaw) {
            $datumClean = ([datetime]$datumRaw).ToString("yyyy-MM-dd_HH-mm-ss")
        }

        $imgIndex = 1
        foreach ($img in $p.data.embed.images) {
            $blobCid = $img.image.ref.'$link'
            
            if ($blobCid) {
                $dateiname = "$datumClean`_$imgIndex.jpg"
                
                # XRPC-Download direkt ueber eurosky.social
                $downloadUrl = "https://$pdsHost/xrpc/com.atproto.sync.getBlob?did=$userDid&cid=$blobCid"
                
                [PSCustomObject]@{
                    URL  = $downloadUrl
                    Pfad = Join-Path $medienOrdner $dateiname
                }
                $imgIndex++
            }
        }
    }
}

Write-Host "Gefundene Bilder: $($medien.Count). Starte Parallel-Download von $pdsHost..." -ForegroundColor Yellow

if ($medien.Count -gt 0) {
    $medien | ForEach-Object -Parallel {
        $item = $_
        if (-not (Test-Path $item.Pfad)) {
            try {
                Invoke-WebRequest -Uri $item.URL -OutFile $item.Pfad -ErrorAction Stop
                Write-Host "Geladen: $($item.Pfad)" -ForegroundColor Green
            } catch {
                $errText = $_.Exception.Message
                Write-Host "Fehler bei $($item.URL) -> $errText" -ForegroundColor Red
            }
        } else {
            Write-Host "Bereits vorhanden: $($item.Pfad)" -ForegroundColor DarkGray
        }
    } -ThrottleLimit 10

    Write-Host "Fertig! Alle Bilder liegen im Ordner '$medienOrdner'." -ForegroundColor Green
}