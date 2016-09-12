# Maintainer: Egor Kovetskiy <e.kovetskiy@office.ngs.ru>
pkgname=nucleus
pkgver=${PKGVER:-autogenerated}
pkgrel=${PKGREL:-1}
pkgdesc="tokens distribution service"
arch=('i686' 'x86_64')
license=('GPL')
depends=(
)
makedepends=(
	'go'
	'git'
)

source=(
	"nucleus::git://github.com/kovetskiy/nucleus.git#branch=${BRANCH:-master}"
)

md5sums=(
	'SKIP'
)

backup=(
    "etc/nucleus/nucleus.conf"
)

pkgver() {
	if [[ "$PKGVER" ]]; then
		echo "$PKGVER"
		return
	fi

	cd "$srcdir/$pkgname"
	local date=$(git log -1 --format="%cd" --date=short | sed s/-//g)
	local count=$(git rev-list --count HEAD)
	local commit=$(git rev-parse --short HEAD)
	echo "$date.${count}_$commit"
}

build() {
	cd "$srcdir/$pkgname"

	if [ -L "$srcdir/$pkgname" ]; then
		rm "$srcdir/$pkgname" -rf
		mv "$srcdir/.go/src/$pkgname/" "$srcdir/$pkgname"
	fi

	rm -rf "$srcdir/.go/src"

	mkdir -p "$srcdir/.go/src"

	export GOPATH="$srcdir/.go"

	mv "$srcdir/$pkgname" "$srcdir/.go/src/"

	cd "$srcdir/.go/src/$pkgname/"
	ln -sf "$srcdir/.go/src/$pkgname/" "$srcdir/$pkgname"

	git submodule update --init

	go get -v \
		-gcflags "-trimpath $GOPATH/src" \
		-ldflags="-X main.version=$pkgver-$pkgrel"
}

package() {
	find "$srcdir/.go/bin/" -type f -executable | while read filename; do
		install -DT "$filename" "$pkgdir/usr/bin/$(basename $filename)"
	done
	install -DT -m0755 \
        "$srcdir/nucleus.service" \
        "$pkgdir/usr/lib/systemd/system/nucleus.service"
	install -DT -m0600 \
        "$srcdir/conf/nucleus.conf" \
        "$pkgdir/etc/nucleus/nucleus.conf"
}
