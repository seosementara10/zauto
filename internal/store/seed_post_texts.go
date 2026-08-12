package store

import (
	"context"
	"fmt"
)

var defaultPostTexts = map[string][]string{
	PostTextCategoryPersonal: {
		"Selamat pagi! Semoga hari ini penuh berkah dan kebahagiaan.",
		"Halo semuanya, senang bisa berbagi momen indah di hari yang cerah ini.",
		"Assalamualaikum, semoga kita semua diberi kesehatan dan rezeki yang melimpah.",
		"Good morning! Yuk mulai hari dengan semangat positif dan penuh syukur.",
		"Hai teman-teman, jangan lupa tersenyum hari ini - senyummu itu berharga.",
		"Selamat siang! Semoga aktivitas kita hari ini berjalan lancar dan menyenangkan.",
		"Halo! Terima kasih sudah menjadi bagian dari perjalanan hidupku.",
		"Semoga hari Jumat penuh keberkahan dan menjadi penutup minggu yang indah.",
		"Selamat malam, semoga istirahat malam ini membawa ketenangan dan mimpi indah.",
		"Hai! Jangan menyerah - setiap langkah kecil tetap membawa kita lebih dekat ke tujuan.",
		"Semangat pagi! Hari baru, peluang baru, dan harapan baru menanti kita.",
		"Terima kasih Tuhan atas nikmat hari ini. Semoga kita selalu bersyukur.",
		"Halo! Semoga cuaca hati kita selalu cerah seperti langit pagi ini.",
		"Selamat beraktivitas! Jaga kesehatan, jaga semangat, dan jaga kebaikan hati.",
		"Hai semuanya! Ingatlah untuk berbuat baik, sekecil apapun itu.",
		"Semoga akhir pekan ini membawa waktu berkualitas bersama orang tersayang.",
		"Hello! Spread kindness today - dunia butuh lebih banyak kebaikan.",
		"Selamat pagi, sahabat! Semoga rezeki dan kebahagiaan selalu menemanimu.",
		"Hai! Tetap semangat ya, kamu lebih kuat dari yang kamu kira.",
		"Semoga hari ini membawa kabar baik, tawa riang, dan hati yang tenang.",
	},
	PostTextCategoryFanpage: {
		"Halo sahabat fanpage! Terima kasih sudah setia menyimak update kami.",
		"Selamat pagi! Fanpage kami siap menemani hari Anda dengan informasi bermanfaat.",
		"Terima kasih atas dukungan dan interaksi positif dari semua pengikut setia.",
		"Hai! Jangan lewatkan update menarik dari fanpage kami hari ini.",
		"Semoga hari Anda produktif - kami siap berbagi tips dan inspirasi.",
		"Hello followers! Terima kasih sudah menjadi bagian dari komunitas kami.",
		"Selamat siang! Simak informasi terbaru yang kami bagikan khusus untuk Anda.",
		"Fanpage ini hadir untuk memberi manfaat, inspirasi, dan kabar terkini.",
		"Terima kasih sudah like, share, dan komentar - dukungan Anda sangat berarti.",
		"Halo! Nantikan konten menarik dari kami sepanjang minggu ini.",
		"Semoga hari ini membawa keberkahan bagi bisnis dan komunitas kita bersama.",
		"Update pagi: kami siap melayani dan merespons pertanyaan Anda dengan senang hati.",
		"Terima kasih sudah mempercayai fanpage kami sebagai sumber informasi.",
		"Hai sahabat! Yuk jaga interaksi positif dan saling mendukung di sini.",
		"Selamat datang di fanpage kami - tempat berbagi informasi dan semangat.",
		"Kami menghargai setiap masukan dari Anda untuk terus berkembang.",
		"Good morning! Stay tuned for today's updates and announcements.",
		"Semoga konten fanpage ini bermanfaat dan menginspirasi hari Anda.",
		"Terima kasih sudah follow - mari tumbuh dan berkembang bersama.",
		"Halo! Fanpage kami buka untuk saran, pertanyaan, dan kolaborasi positif.",
	},
	PostTextCategoryGroup: {
		"Halo teman-teman grup! Semoga semua dalam keadaan baik dan penuh semangat.",
		"Selamat pagi anggota grup! Mari jaga kehangatan dan saling menghargai.",
		"Assalamualaikum wr wb, semoga kita semua diberi kemudahan dalam setiap urusan.",
		"Hai semuanya! Terima kasih sudah aktif dan positif di grup ini.",
		"Selamat siang! Yuk saling berbagi informasi bermanfaat di grup kita.",
		"Halo grup! Ingatkan satu sama lain dengan kebaikan dan kata-kata positif.",
		"Semoga hari ini membawa keberkahan bagi kita semua di grup ini.",
		"Terima kasih admin dan member yang selalu menjaga kondusivitas grup.",
		"Hai! Mari kita jadikan grup ini tempat belajar dan saling mendukung.",
		"Selamat malam, semoga istirahat kita tenang dan esok lebih baik.",
		"Halo teman! Jangan ragu berbagi ide dan pengalaman di grup ini.",
		"Semangat pagi! Grup ini semakin hidup karena partisipasi aktif Anda.",
		"Terima kasih sudah menghormati aturan grup — mari jaga lingkungan positif.",
		"Hai anggota grup! Semoga rezeki dan kebahagiaan selalu menyertai kita.",
		"Selamat beraktivitas! Saling remind dengan hal baik ya, teman-teman.",
		"Halo! Grup ini terbuka untuk diskusi sehat dan saling membantu.",
		"Semoga akhir pekan kita penuh kebersamaan dan kegiatan yang menyenangkan.",
		"Terima kasih sudah menjadi bagian komunitas yang solid dan positif.",
		"Hai semuanya! Yuk jaga sopan santun dan saling menghargai pendapat.",
		"Selamat pagi grup! Semoga hari ini penuh kebaikan dan kebermanfaatan.",
	},
}

// SeedDefaultPostTexts inserts 20 default greeting texts per category when empty.
func (s *Store) SeedDefaultPostTexts(ctx context.Context) error {
	for category, lines := range defaultPostTexts {
		n, err := s.CountPostTexts(ctx, category)
		if err != nil {
			return fmt.Errorf("count %s: %w", category, err)
		}
		if n > 0 {
			continue
		}
		for _, body := range lines {
			if _, err := s.CreatePostText(ctx, category, body, ""); err != nil {
				return fmt.Errorf("seed %s: %w", category, err)
			}
		}
	}
	return nil
}
