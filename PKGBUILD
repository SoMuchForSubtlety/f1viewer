# Maintainer: SoMuchForSubtlety <s0muchfrsubtlety@gmail.com>
pkgname=f1viewer
pkgver=1.1.0
pkgrel=1
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
sha256sums=('3d74093d54000c65f46a8a11aa55a5b535367689f963adace17d0f4c0ddf7802')

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
