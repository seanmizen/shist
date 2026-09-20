# shist - Sean's History Tool

https://github.com/user-attachments/assets/c2d6e4dc-fdf3-4b32-a70b-20f771d4eda3

### Install

Download the archive for your platform from the
[latest release](https://github.com/seanmizen/shist/releases/latest), then:

```bash
tar -xzf shist_*_darwin_arm64.tar.gz
sudo mv shist /usr/local/bin/
shist
```

Windows: unzip and put `shist.exe` somewhere on your `PATH`.

Every release ships a `checksums.txt` if you want to verify the download:

```bash
shasum -a 256 -c checksums.txt --ignore-missing
```

### Build from source

_requires golang_

```bash
git clone https://github.com/seanmizen/shist
cd shist
make install
shist
```

### Releasing

Tag and push - CI builds all six platform archives and publishes them:

```bash
git tag v1.0.0
git push origin v1.0.0
```

`make release` does the same build locally, into `dist/`.

### Licence

MIT - see [LICENSE](LICENSE).
