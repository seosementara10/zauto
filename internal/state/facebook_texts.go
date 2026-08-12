package state

// Facebook UI text catalogs — single source for detector scoring and action engine.

var (
	SaveLoginTitleTexts = []string{
		"Simpan info login Anda?", "Simpan info login", "Save your login", "Save login info",
		"Simpan info login ke Facebook",
	}
	SaveLoginSaveTexts = []string{"Simpan", "Save"}
	SaveLoginLaterTexts = []string{
		"LAIN KALI", "Lain Kali", "Lain kali", "Not now", "Not Now",
		"Maybe later", "Maybe Later", "Skip", "Nanti saja",
	}

	ContactFollowTitleTexts = []string{
		"akses kontak menemukan orang", "find people to follow",
		"Izinkan akses kontak", "contact access",
	}
	ContactSkipConfirmTitleTexts = []string{
		"Yakin tidak ingin mengunggah kontak",
		"don't want to upload your contacts",
		"sure you don't want to upload",
	}
	ContactFollowSkipTexts = []string{"Lewati", "Skip", "SKIP", "Lewati saja"}
	PostLoginSkipTexts     = ContactFollowSkipTexts

	LocationServicesIntroTexts = []string{
		"Layanan Lokasi", "Location Services",
		"Untuk menggunakan Layanan Lokasi",
		"izinkan Facebook untuk mengakses lokasi",
		"allow Facebook to access your location",
	}
	ContactFollowContinueTexts = []string{"Lanjutkan", "Continue", "Next", "Berikutnya"}
	ContactUploadTexts         = []string{"Unggah kontak", "Upload contacts", "Upload Contacts"}

	LogoutConfirmTitleTexts = []string{
		"Logout dari akun Anda?", "Logout dari akun Anda",
		"Log out of your account", "Keluar dari akun",
	}
	LogoutConfirmButtonTexts = []string{"LOGOUT", "Log out", "LOG OUT", "Logout"}
	LogoutMenuItemTexts      = []string{"Keluar", "KELUAR", "Log out", "Logout"}
	LogoutConfirmTexts       = append(append([]string(nil), LogoutConfirmButtonTexts...), LogoutMenuItemTexts...)
	LogoutCancelTexts        = []string{"BATALKAN", "Batalkan", "Cancel", "CANCEL"}

	LoggedInFeedHints = []string{
		"What's on your mind", "Apa yang Anda pikirkan",
		"Suka", "Komentar", "Bagikan", "Menu",
	}

	// FeedComposerTexts — tap target to open create-post composer on home feed.
	FeedComposerTexts = []string{
		"What's on your mind?", "What's on your mind",
		"Apa yang Anda pikirkan?", "Apa yang Anda pikirkan",
		"Write something", "Tulis sesuatu",
	}
	FeedComposerScreenTexts = []string{
		"Create post", "Buat postingan", "Create Post", "Buat Postingan",
		"Postingan baru", "New post", "New Post",
	}
	PostComposerNextTexts = []string{
		"Berikutnya", "Next",
	}
	PostComposerNewTitleTexts = []string{
		"Postingan baru", "New post", "New Post",
	}
	// PostFinalPublishTexts — final confirm on review/settings screen (exact match only in engine).
	PostFinalPublishTexts = []string{
		"POST", "PUBLISH", "Kirim", "Publish", "Posting",
		"BERBAGI", "Bagikan", "Share", "SHARE",
	}
	PostSettingsScreenTexts = []string{
		"Pengaturan postingan", "Post settings", "Post Settings",
	}
	PostPhotoButtonTexts = []string{
		"Photo", "Foto", "Gambar", "Add photos", "Tambahkan foto",
	}
	// PostPublishButtonTexts kept for detector hints; post engine uses PostFinalPublishTexts.
	PostPublishButtonTexts = []string{
		"POST", "Post", "Posting", "PUBLISH", "Publish", "Bagikan", "Kirim",
	}
	GalleryRecentTexts = []string{
		"Recent", "Terbaru", "Gallery", "Galeri", "Photos", "Foto",
	}
	GalleryImagePickerTexts = []string{
		"Choose photo", "Pilih foto", "Select photo", "Select image",
	}

	SwitchProfileTexts = []string{
		"Switch profile", "Switch profiles", "Ganti profil", "Ganti Profil",
		"Switch account", "Ganti akun", "Switch to", "Beralih ke",
	}
	SeeAllProfilesTexts = []string{
		"See all profiles", "Lihat semua profil", "See all accounts", "Lihat semua akun",
	}
	FanpageFeedHints = []string{
		"Manage Page", "Kelola Halaman", "Promote", "Promosikan", "Professional dashboard",
		"Page", "Halaman",
	}
	FanpageHomeIntroTitleTexts = []string{
		"Memperkenalkan Beranda Khusus",
		"Introducing a Special Home",
		"Beranda Khusus untuk",
		"Special Home for",
	}
	FanpageHomeIntroBodyTexts = []string{
		"Berinteraksi sebagai Halaman",
		"Interact as your Page",
		"ruang khusus yang terpisah",
		"dedicated space separate",
	}
	FanpageHomeIntroSkipTexts = []string{"Lewati", "Skip", "SKIP"}

	PostPromoteTitleTexts = []string{
		"Tingkatkan jangkauan Anda",
		"Boost your reach",
		"Increase your reach",
	}
	PostPromoteButtonTexts = []string{
		"Promosikan postingan",
		"Promote post",
		"Boost post",
		"Promote Post",
	}
	PostPromoteLaterTexts = []string{
		"Lain Kali", "Lain kali", "LAIN KALI",
		"Not now", "Not Now", "Maybe later", "Maybe Later",
	}
	PostPromoteDontShowTexts = []string{
		"Jangan tampilkan pesan ini",
		"Don't show this message",
		"don't show this again",
		"Do not show this message",
	}

	SavedProfileContinueTexts = []string{"Lanjut", "Continue", "Log in"}
	SavedProfileOtherTexts    = []string{
		"Gunakan profil lain", "Use another profile", "Use another account",
	}
	SavedProfileCreateTexts = []string{"Buat akun baru", "Create new account", "Create New Account"}
	SavedProfileMenuContentDescs = []string{
		"More options", "More", "Opsi lainnya", "Menu", "Profile settings", "Pengaturan profil",
	}
	RemoveProfileFromDeviceTexts = []string{
		"Hapus Profil dari perangkat ini", "Hapus profil dari perangkat ini",
		"Remove profile from this device", "Remove from this device",
	}
	RemoveProfileConfirmTexts = []string{
		"HAPUS", "Hapus", "Remove", "REMOVE", "Ya", "Yes", "OK", "OKE",
	}

	LoginEmailFieldTexts = []string{
		"Nomor ponsel atau email", "Mobile number or email", "Email or phone number",
	}
	LoginPasswordFieldTexts = []string{"Kata sandi", "Password"}
	LoginButtonTexts        = []string{"Log in", "Login", "Masuk", "Log In"}
	LoginAccountFinderTexts = []string{
		"Cari akun saya", "Find my account", "Find My Account",
	}
	LoginAccountFinderPromptTexts = []string{
		"Masukkan email atau nomor ponsel Anda untuk login",
		"Enter your email or phone number to log in",
	}
	LoginAccountFinderDismissTexts = []string{"OK", "Ok", "OKE", "Batal", "Cancel", "BATALKAN"}

	KeyboardSettingsTitleTexts = []string{"Setelan", "Settings"}
	KeyboardSettingsNavTexts   = []string{"Kembali ke atas", "Navigate up", "Back"}
	KeyboardSettingsMenuTexts  = []string{"Bahasa", "Language", "Preferensi", "Preferences", "Tema", "Theme"}

	TwoFactorTitleTexts = []string{
		"Two-factor authentication", "Autentikasi dua faktor", "Two factor authentication",
		"Enter login code", "Masukkan kode login", "Enter the login code",
		"Check your notifications", "Periksa notifikasi Anda", "Approve from another device",
		"Confirm it's you", "Konfirmasi bahwa ini Anda", "Login approval needed",
	}
	TwoFactorActionTexts = []string{
		"Try another way", "Coba cara lain", "Continue", "Lanjutkan", "Submit", "Kirim",
	}

	OnboardingHaveAccountTexts = []string{
		"Saya sudah punya profil", "I already have an account",
	}

	PasswordManagerTitleTexts = []string{
		"Pengelola Sandi Google", "Google Password Manager", "Password Manager",
	}
	PasswordManagerSaveTitleTexts = []string{
		"Simpan sandi ke Pengelola Sandi Google?", "Simpan sandi untuk login ke Facebook?",
		"Save password for login to Facebook?", "Save password to sign in to Facebook?",
	}
	PasswordManagerNeverTexts = []string{
		"Tidak pernah", "Never", "Not now", "Not Now",
	}
	PasswordManagerContinueTexts = []string{
		"Lanjutkan", "Continue", "Selanjutnya", "Next",
	}
	// PasswordManagerCloseContentDescs — X / close on Google Password Manager bottom sheet.
	PasswordManagerCloseContentDescs = []string{
		"Close", "Tutup", "Dismiss", "Close sheet", "Tutup sheet", "Navigate up",
		"Tutup dialog", "Close dialog",
	}

	// OverlayCloseContentDescs — semantic close/dismiss for bottom sheets and dialogs.
	OverlayCloseContentDescs = []string{
		"Close", "Tutup", "Dismiss", "Navigate up", "Close sheet",
	}
	OverlayCloseTexts = []string{"×", "✕"}

	LocaleSetupMessageTexts = []string{
		"trouble setting up Facebook in Indonesia",
		"having trouble setting up Facebook in Indonesia",
		"Kesulitan menyiapkan Facebook di Indonesia",
	}
	LocaleSetupTryAgainTexts = []string{"Try again", "Coba lagi"}
	LocaleSetupContinueEnglishTexts = []string{
		"Continue in English (US)", "Continue in English",
		"Lanjutkan dalam Bahasa Inggris", "Lanjutkan dalam bahasa Inggris",
	}

	NotificationPermissionPromptTexts = []string{
		"mengirim notifikasi", "send you notifications", "send notifications",
		"notifikasi kepada Anda", "notifications to you", "kirim notifikasi",
	}
	ContactPermissionPromptTexts = []string{
		"access your contacts", "read your contacts", "contacts on your phone",
		"allow Facebook to access your contacts", "allow access to contacts",
		"mengakses kontak", "akses ke kontak Anda", "baca kontak Anda",
	}
	LocationPermissionPromptTexts = []string{
		"access your location", "akses lokasi", "location of this device",
		"precise location", "lokasi perangkat",
	}
	CameraPermissionPromptTexts = []string{
		"take pictures", "record video", "access the camera", "akses kamera",
	}
	PermissionDenyTexts = []string{
		"Don't allow", "Don't Allow", "Jangan izinkan", "JANGAN IZINKAN", "Tolak", "Deny",
	}
	PermissionAllowTexts = []string{
		"Allow", "Izinkan", "While using the app", "Only this time",
	}
)
