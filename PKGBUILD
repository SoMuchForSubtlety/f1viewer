# Maintainer: SoMuchForSubtlety <s0muchfrsubtlety@gmail.com>
pkgname=f1viewer
pkgver=1.0.0
pkgrel=2
pkgdesc="TUI client for F1TV"
arch=('x86_64')
url="https://github.com/SoMuchForSubtlety/f1viewer"
license=('GPL3')
depends=('mpv')
optdepends=('xclip: copying URLs to clipboard'
            'keepassxc: secret store backend'
            'pass: secret store backend'
            'gnome-keyring: secret store backend'
            'kwallet: secret store backend')
makedepends=('go-pie')
source=("${pkgname}-${pkgver}.tar.gz::https://github.com/SoMuchForSubtlety/f1viewer/archive/v${pkgver}.tar.gz")
sha256sums=('40986b7ed358adf2299882318d9ca81ef7f1bf730ed30b3c5da362518987d6af')

prepare() {
  if pacman -Qi keepassxc >/dev/null 2>&1; then
    return 0
  elif pacman -Qi pass >/dev/null 2>&1; then
    return 0
  elif pacman -Qi gnome-keyring >/dev/null 2>&1; then
    return 0
  elif pacman -Qi kwallet >/dev/null 2>&1; then
    return 0
  fi
  
  echo 'You need to install a secrets backend like gnome-keyring, kwallet, pass or keepassxc - you might want to install it with pacman -S --asdeps package_name' >&2;
  return 1
}

build() {
  cd "${pkgname}-${pkgver}"
  go build \
    -trimpath \
    -ldflags="-extldflags ${LDFLAGS} -s -w -X main.version=${pkgver}" \
    -o $pkgname .
}

package() {
  cd $pkgname-$pkgver
  install -Dm755 $pkgname "$pkgdir"/usr/bin/$pkgname
}
