# Maintainer: SoMuchForSubtlety <s0muchfrsubtlety@gmail.com>
pkgname=f1viewer
pkgver=1.3.0
pkgrel=1
pkgdesc="TUI client for F1TV"
arch=('x86_64')
url="https://github.com/SoMuchForSubtlety/f1viewer"
license=('GPL3')
optdepends=('mpv: play videos using mpv'
            'xclip: copying URLs to clipboard'
            'keepassxc: secret store backend'
            'pass: secret store backend'
            'gnome-keyring: secret store backend'
            'kwallet: secret store backend')
makedepends=('go-pie')
source=("${pkgname}-${pkgver}.tar.gz::https://github.com/SoMuchForSubtlety/f1viewer/archive/v${pkgver}.tar.gz")
sha256sums=('26837bc1a78395698eeecde4fe3a22b89b4d87b4210a409530a72a718114fb2a')

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
