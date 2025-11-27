module github.com/xaionaro-go/player

go 1.24.4

toolchain go1.24.7

replace github.com/asticode/go-astiav v0.36.0 => github.com/xaionaro-go/astiav v0.0.0-20251114192847-048826e6dc3a

replace github.com/dexterlb/mpvipc => github.com/xaionaro-go/mpvipc v0.0.0-20251019230357-e0f534e5dde4

require (
	fyne.io/fyne/v2 v2.7.0
	github.com/AgustinSRG/go-child-process-manager v1.0.1
	github.com/adrg/libvlc-go/v3 v3.1.6
	github.com/asticode/go-astiav v0.36.0
	github.com/blang/mpv v0.0.0-20160810175505-d56d7352e068
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/dexterlb/mpvipc v0.0.0-00010101000000-000000000000
	github.com/facebookincubator/go-belt v0.0.0-20250308011339-62fb7027b11f
	github.com/goccy/go-yaml v1.18.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/spf13/pflag v1.0.10
	github.com/xaionaro-go/audio v0.0.0-20250426140416-6a9b3f1c8737
	github.com/xaionaro-go/avpipeline v0.0.0-20251127190533-54f212ffa588
	github.com/xaionaro-go/datacounter v1.0.4
	github.com/xaionaro-go/logwriter v0.0.0-20250111154941-c3f7a1a2d567
	github.com/xaionaro-go/observability v0.0.0-20251102143534-3aeb2a25e57d
	github.com/xaionaro-go/secret v0.0.0-20250111141743-ced12e1082c2
	github.com/xaionaro-go/xcontext v0.0.0-20250111150717-e70e1f5b299c
	github.com/xaionaro-go/xfyne v0.0.0-20250615190411-4c96281f6e25
	github.com/xaionaro-go/xpath v0.0.0-20250111145115-55f5728f643f
	github.com/xaionaro-go/xsync v0.0.0-20250928140805-f801683b71ba
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/ebitengine/gomobile v0.0.0-20250923094054-ea854a63cce1 // indirect
	github.com/ebitengine/hideconsole v1.0.0 // indirect
	github.com/go-gst/go-glib v1.4.0 // indirect
	github.com/jezek/xgb v1.1.1 // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
	github.com/samber/lo v1.52.0 // indirect
	github.com/xaionaro-go/rpn v0.0.0-20250818130635-1419b5218722 // indirect
	golang.org/x/sync v0.18.0 // indirect
	tailscale.com v1.86.5 // indirect
)

require (
	codeberg.org/go-fonts/liberation v0.5.0 // indirect
	codeberg.org/go-latex/latex v0.1.0 // indirect
	codeberg.org/go-pdf/fpdf v0.10.0 // indirect
	fyne.io/systray v1.11.1-0.20250603113521-ca66a66d8b58 // indirect
	git.sr.ht/~sbinet/gg v0.6.0 // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/DataDog/gostackparse v0.7.0 // indirect
	github.com/ajstarks/svgo v0.0.0-20211024235047-1546f124cd8b // indirect
	github.com/asticode/go-astikit v0.55.0 // indirect
	github.com/av-elier/go-decimal-to-rational v0.0.0-20250603203441-f39a07f43ff3 // indirect
	github.com/campoy/embedmd v1.0.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/ebitengine/oto/v3 v3.4.0 // indirect
	github.com/ebitengine/purego v0.9.0 // indirect
	github.com/fredbi/uri v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fyne-io/gl-js v0.2.0 // indirect
	github.com/fyne-io/glfw-js v0.3.0 // indirect
	github.com/fyne-io/image v0.1.1 // indirect
	github.com/fyne-io/oksvg v0.2.0 // indirect
	github.com/go-gl/gl v0.0.0-20231021071112-07e5d0ea2e71 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20240506104042-037f3cc74f2a // indirect
	github.com/go-gst/go-gst v1.4.0
	github.com/go-ng/container v0.0.0-20220615121757-4740bf4bbc52 // indirect
	github.com/go-ng/slices v0.0.0-20230703171042-6195d35636a2 // indirect
	github.com/go-ng/sort v0.0.0-20220617173827-2cc7cd04f7c7 // indirect
	github.com/go-ng/xatomic v0.0.0-20251124145245-9a7a1838d3aa // indirect
	github.com/go-ng/xmath v0.0.0-20230704233441-028f5ea62335 // indirect
	github.com/go-ng/xsort v0.0.0-20250330112557-d2ee7f01661c // indirect
	github.com/go-text/render v0.2.0 // indirect
	github.com/go-text/typesetting v0.3.0 // indirect
	github.com/godbus/dbus/v5 v5.1.1-0.20230522191255-76236955d466 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hack-pad/go-indexeddb v0.3.2 // indirect
	github.com/hack-pad/safejs v0.1.0 // indirect
	github.com/hajimehoshi/ebiten/v2 v2.9.4
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/huandu/go-tls v1.0.1 // indirect
	github.com/jeandeaual/go-locale v0.0.0-20250612000132-0ef82f21eade // indirect
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	github.com/jsummers/gobmp v0.0.0-20230614200233-a9de23ed2e25 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/lmpizarro/go_ehlers_indicators v0.0.0-20220405041400-fd6ced57cf1a // indirect
	github.com/montanaflynn/stats v0.6.6 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/nicksnyder/go-i18n/v2 v2.5.1 // indirect
	github.com/phuslu/goid v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rymdport/portal v0.4.2 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/srwiley/oksvg v0.0.0-20221011165216-be6e8873101c // indirect
	github.com/srwiley/rasterx v0.0.0-20220730225603-2ab79fcdd4ef // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/xaionaro-go/androidetc v0.0.0-20250824193302-b7ecebb3b825 // indirect
	github.com/xaionaro-go/avcommon v0.0.0-20250823173020-6a2bb1e1f59d // indirect
	github.com/xaionaro-go/avmediacodec v0.0.0-20250505012527-c819676502d8 // indirect
	github.com/xaionaro-go/gorex v0.0.0-20241010205749-bcd59d639c4d // indirect
	github.com/xaionaro-go/libsrt v0.0.0-20250505013920-61d894a3b7e9 // indirect
	github.com/xaionaro-go/logrustash v0.0.0-20240804141650-d48034780a5f // indirect
	github.com/xaionaro-go/ndk v0.0.0-20250420195304-361bb98583bf // indirect
	github.com/xaionaro-go/object v0.0.0-20241026212449-753ce10ec94c // indirect
	github.com/xaionaro-go/proxy v0.0.0-20250525144747-579f5a891c15 // indirect
	github.com/xaionaro-go/sockopt v0.0.0-20250823181757-5c02c9cd7b51 // indirect
	github.com/xaionaro-go/spinlock v0.0.0-20200518175509-30e6d1ce68a1 // indirect
	github.com/xaionaro-go/typing v0.0.0-20221123235249-2229101d38ba // indirect
	github.com/xaionaro-go/unsafetools v0.0.0-20241024014258-a46e1ce3763e // indirect
	github.com/yuin/goldmark v1.7.8 // indirect
	gocv.io/x/gocv v0.41.0 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/exp v0.0.0-20250813145105-42675adae3e6 // indirect
	golang.org/x/image v0.31.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	gonum.org/v1/plot v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/blake3 v1.4.0 // indirect
)
