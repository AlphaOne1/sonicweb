// SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

package dirindex

// Translation represents a localized version of a resource with fields for its name, size,
// and modification information.
type Translation struct {
	ListingName  string
	Name         string
	Size         string
	LastModified string
	RTL          bool
}

// Translations is a map where keys are language codes and values are Translation structs containing localized text.
//
//nolint:gochecknoglobals,goconst,gosmopolitan,lll,misspell
var Translations = map[string]Translation{
	"af":  {ListingName: "Gidsinhoud", Name: "Naam", Size: "Grootte", LastModified: "Laas gewysig"},                                 // Afrikaans
	"akk": {ListingName: "𒁾𒂍 𒈬", Name: "𒈬", Size: "𒈠𒁕𒁺", LastModified: "𒌓"},                                                         // Akkadian
	"am":  {ListingName: "የማውጫ ዝርዝር", Name: "ስም", Size: "መጠን", LastModified: "መጨረሻ የተሻሻለው"},                                         // Amharic
	"ar":  {ListingName: "قائمة الدليل", Name: "الاسم", Size: "الحجم", LastModified: "آخر تعديل", RTL: true},                        // Arabic
	"arc": {ListingName: "ܡܢܝܢܐ ܕܡܕܒܪܢܐ", Name: "ܫܡܐ", Size: "ܡܫܘܚܬܐ", LastModified: "ܫܘܚܠܦܐ ܐܚܪܝܐ", RTL: true},                     // Aramaic
	"as":  {ListingName: "ডাইৰেকটৰী তালিকা", Name: "নাম", Size: "আকাৰ", LastModified: "অন্তিম সংশোধন"},                              // Assamese
	"ay":  {ListingName: "Suti uñacht'äwi", Name: "Suti", Size: "Jach'a kay", LastModified: "Qhipa mayjt'ayata"},                    // Aymara
	"az":  {ListingName: "Kataloq siyahısı", Name: "Ad", Size: "Ölçü", LastModified: "Son dəyişiklik"},                              // Azerbaijani
	"be":  {ListingName: "Спіс каталога", Name: "Імя", Size: "Памер", LastModified: "Апошняя змена"},                                // Belarusian
	"ber": {ListingName: "ⵜⴰⴱⴷⴰⵔⵜ ⵏ ⵓⴽⴰⵔⴰⵎ", Name: "ⵉⵙⵎ", Size: "ⵜⴰⴽⵯⵜⴰ", LastModified: "ⴰⵙⵏⴼⵍ ⴰⵏⴳⴳⴰⵔⵓ"},                            // Berber
	"bg":  {ListingName: "Списък на директорията", Name: "Име", Size: "Размер", LastModified: "Последна промяна"},                   // Bulgarian
	"bn":  {ListingName: "ডিরেক্টরি তালিকা", Name: "নাম", Size: "আকার", LastModified: "শেষ পরিবর্তন"},                               // Bengali
	"br":  {ListingName: "Roll ar restr", Name: "Anv", Size: "Ment", LastModified: "Kemm diwezhañ"},                                 // Breton
	"bs":  {ListingName: "Popis direktorija", Name: "Naziv", Size: "Veličina", LastModified: "Posljednja izmjena"},                  // Bosnian
	"ca":  {ListingName: "Llistat del directori", Name: "Nom", Size: "Mida", LastModified: "Última modificació"},                    // Catalan
	"chr": {ListingName: "ᏗᎧᏃᏗᏍᎩ ᏗᎦᏛ", Name: "ᏗᎦᏙᎥ", Size: "ᎠᏍᏓᏅᏅ", LastModified: "ᎤᏓᏡᎲᏍᏓᏅᏅ ᎤᏩᏓᏛ"},                                  // Cherokee
	"cop": {ListingName: "ⲡⲓⲕⲁⲧⲁⲗⲟⲅⲟⲥ ⲛ̀ⲧⲉ ⲡⲓⲙⲁ", Name: "ⲓⲛⲟⲩⲙ", Size: "ⲙⲉϣⲓ", LastModified: "ϣⲓⲃⲉ ⲙⲁⲉ"},                            // Coptic
	"cs":  {ListingName: "Výpis adresáře", Name: "Název", Size: "Velikost", LastModified: "Poslední změna"},                         // Czech
	"cy":  {ListingName: "Rhestr cyfeiriadur", Name: "Enw", Size: "Maint", LastModified: "Wedi'i addasu ddiwethaf"},                 // Welsh
	"da":  {ListingName: "Mappeliste", Name: "Navn", Size: "Størrelse", LastModified: "Sidst ændret"},                               // Danish
	"de":  {ListingName: "Verzeichnisinhalt", Name: "Name", Size: "Größe", LastModified: "Letzte Änderung"},                         // German
	"dsb": {ListingName: "Lisćina zapisa", Name: "Mě", Size: "Wjelikosć", LastModified: "Slědna změna"},                             // Lower Sorbian
	"egy": {ListingName: "𓏃𓏏 𓂋𓈖 𓉐", Name: "𓂋𓈖", Size: "𓉻𓂠", LastModified: "𓆎𓈖𓏏𓇳"},                                                   // Ancient Egyptian
	"el":  {ListingName: "Λίστα καταλόγου", Name: "Όνομα", Size: "Μέγεθος", LastModified: "Τελευταία τροποποίηση"},                  // Greek
	"en":  {ListingName: "Directory Listing", Name: "Name", Size: "Size", LastModified: "Last Modified"},                            // English (default, must remain!)
	"eo":  {ListingName: "Dosieruja listo", Name: "Nomo", Size: "Grando", LastModified: "Lasta modifo"},                             // Esperanto
	"es":  {ListingName: "Listado de directorio", Name: "Nombre", Size: "Tamaño", LastModified: "Última modificación"},              // Spanish
	"et":  {ListingName: "Kataloogi loend", Name: "Nimi", Size: "Suurus", LastModified: "Viimati muudetud"},                         // Estonian
	"etp": {ListingName: "𐌄𐌔𐌀𐌍𐌔𐌉𐌄", Name: "𐌄𐌔𐌀", Size: "𐌐𐌄𐌄𐌋", LastModified: "𐌛𐌉𐌋 𐌋𐌖𐌐"},                                             // Etruscan
	"eu":  {ListingName: "Direktorio-zerrenda", Name: "Izena", Size: "Tamaina", LastModified: "Azken aldaketa"},                     // Basque
	"fa":  {ListingName: "فهرست راهنما", Name: "نام", Size: "اندازه", LastModified: "آخرین تغییرات", RTL: true},                     // Persian
	"ff":  {ListingName: "Doggu", Name: "Innde", Size: "Mawnudi", LastModified: "Waylaama sakketande"},                              // Fulah
	"fi":  {ListingName: "Hakemistoluettelo", Name: "Nimi", Size: "Koko", LastModified: "Muokattu viimeksi"},                        // Finnish
	"fo":  {ListingName: "Yvirlit", Name: "Navn", Size: "Stødd", LastModified: "Seinast broytt"},                                    // Faroese
	"fr":  {ListingName: "Contenu du répertoire", Name: "Nom", Size: "Taille", LastModified: "Dernière modification"},               // French
	"fur": {ListingName: "Lise de cartele", Name: "Non", Size: "Dimension", LastModified: "Ultime modifiche"},                       // Friulian
	"fy":  {ListingName: "Map ynhâld", Name: "Namme", Size: "Grutte", LastModified: "Lêst feroare"},                                 // Western Frisian
	"ga":  {ListingName: "Liosta comhadlainne", Name: "Ainm", Size: "Méid", LastModified: "Modhnaithe go deireanach"},               // Irish
	"gd":  {ListingName: "Liosta pasgain", Name: "Ainm", Size: "Meud", LastModified: "Atharrachadh mu dheireadh"},                   // Scottish Gaelic
	"gem": {ListingName: "ᚹᛁᚴᛁᛏᛁ", Name: "ᚾᚨᛗᛟ", Size: "ᛗᛖᛏᛟ", LastModified: "ᚨᛁᛞᛁ"},                                                // Germanic languages / Proto-Germanic
	"gl":  {ListingName: "Lista do directorio", Name: "Nome", Size: "Tamaño", LastModified: "Última modificación"},                  // Galician
	"gmy": {ListingName: "𐀵𐀟𐀼", Name: "𐀃𐀜𐀔", Size: "𐀕𐀼", LastModified: "𐀀𐀕𐀜"},                                                       // Mycenaean Greek
	"gn":  {ListingName: "Tembiasakue Ñanduti", Name: "Téra", Size: "Tuichakue", LastModified: "Oñemoambue pahápe"},                 // Guaraní
	"gu":  {ListingName: "ડિરેક્ટરી સૂચિ", Name: "નામ", Size: "કદ", LastModified: "છેલ્લે ફેરફાર કરેલ"},                             // Gujarati
	"ha":  {ListingName: "Jerin kundin adireshi", Name: "Suna", Size: "Girma", LastModified: "Gyaran ƙarshe"},                       // Hausa
	"he":  {ListingName: "רשימת תיקייה", Name: "שם", Size: "גודל", LastModified: "שינוי אחרון", RTL: true},                          // Hebrew
	"hi":  {ListingName: "निर्देशिका सूची", Name: "नाम", Size: "आकार", LastModified: "अंतिम परिवर्तन"},                              // Hindi
	"hit": {ListingName: "𒁾𒄭", Name: "𒈬", Size: "𒃲", LastModified: "𒌓"},                                                             // Hittite
	"hr":  {ListingName: "Popis direktorija", Name: "Naziv", Size: "Veličina", LastModified: "Zadnja izmjena"},                      // Croatian
	"hsb": {ListingName: "Lisćina zapisa", Name: "Mjeno", Size: "Wulkosć", LastModified: "Poslednja změna"},                         // Upper Sorbian
	"hu":  {ListingName: "Könyvtárlista", Name: "Név", Size: "Méret", LastModified: "Utolsó módosítás"},                             // Hungarian
	"hy":  {ListingName: "Պանակի ցուցակ", Name: "Անուն", Size: "Չափ", LastModified: "Վերջին փոփոխություն"},                          // Armenian
	"id":  {ListingName: "Daftar direktori", Name: "Nama", Size: "Ukuran", LastModified: "Terakhir diubah"},                         // Indonesian
	"ig":  {ListingName: "Ndepụta ndekọ", Name: "Aha", Size: "Nha", LastModified: "Mgbanwe ikpeazụ"},                                // Igbo
	"is":  {ListingName: "Möppulisti", Name: "Nafn", Size: "Stærð", LastModified: "Síðast breytt"},                                  // Icelandic
	"it":  {ListingName: "Elenco della directory", Name: "Nome", Size: "Dimensione", LastModified: "Ultima modifica"},               // Italian
	"iu":  {ListingName: "ᐊᓪᓚᖁᑎᒃᑯᕕᒃ", Name: "ᐊᑎᖓ", Size: "ᐊᖏᓂᖓ", LastModified: "ᑭᖑᓪᓕᖅᐄᑦ ᐊᓯᔾᔨᖅᑕᐅᔪᑦ"},                                 // Inuktitut
	"ja":  {ListingName: "ファイル一覧", Name: "名前", Size: "サイズ", LastModified: "最終更新"},                                                   // Japanese
	"jv":  {ListingName: "Dhaptar direktori", Name: "Jeneng", Size: "Ukuran", LastModified: "Pungkasan diowahi"},                    // Javanese
	"ka":  {ListingName: "დირექტორიის სია", Name: "სახელი", Size: "ზომა", LastModified: "ბოლო ცვლილება"},                            // Georgian
	"kk":  {ListingName: "Каталог тізімі", Name: "Атауы", Size: "Өлшемі", LastModified: "Соңғы өзгертілген"},                        // Kazakh
	"kl":  {ListingName: "Mappip allattorsimaffia", Name: "Ateq", Size: "Angissusia", LastModified: "Kingullermik allanngortippoq"}, // Greenlandic
	"km":  {ListingName: "បញ្ជីថត", Name: "ឈ្មោះ", Size: "ទំហំ", LastModified: "ការកែប្រែចុងក្រោយ"},                                 // Khmer
	"kn":  {ListingName: "ಡೈರೆಕ್ಟರಿ ಪಟ್ಟಿ", Name: "ಹೆಸರು", Size: "ಗಾತ್ರ", LastModified: "ಕೊನೆಯದಾಗಿ ಮಾರ್ಪಡಿಸಲಾಗಿದೆ"},                 // Kannada
	"ko":  {ListingName: "디렉터리 목록", Name: "이름", Size: "크기", LastModified: "수정된 날짜"},                                                 // Korean
	"ku":  {ListingName: "لیستی پەڕگەدان", Name: "ناو", Size: "قەبارە", LastModified: "دوا دەستکاری", RTL: true},                    // Kurdish
	"ky":  {ListingName: "Каталог тизмеси", Name: "Аты", Size: "Өлчөмү", LastModified: "Акыркы өзгөртүү"},                           // Kyrgyz
	"la":  {ListingName: "Index directorii", Name: "Nomen", Size: "Magnitudo", LastModified: "Ultima mutatio"},                      // Latin
	"lb":  {ListingName: "Dossiersinhalt", Name: "Numm", Size: "Gréisst", LastModified: "Lescht Ännerung"},                          // Luxembourgish
	"lo":  {ListingName: "ລາຍການໄດເຣັກທໍຣີ", Name: "ຊື່", Size: "ຂະໜາດ", LastModified: "ດັດແກ້ຫຼ້າສຸດ"},                             // Lao
	"lt":  {ListingName: "Katalogo sąrašas", Name: "Pavadinimas", Size: "Dydis", LastModified: "Paskutinis pakeitimas"},             // Lithuanian
	"lv":  {ListingName: "Direktorija saraksts", Name: "Nosaukums", Size: "Izmērs", LastModified: "Pēdējās izmaiņas"},               // Latvian
	"mai": {ListingName: "निर्देशिका सूची", Name: "नाम", Size: "आकार", LastModified: "अंतिम परिवर्तन"},                              // Maithili
	"mi":  {ListingName: "Rārangi kōpaki", Name: "Ingoa", Size: "Rahi", LastModified: "Whakarerekē whakamutunga"},                   // Māori
	"mk":  {ListingName: "Список на директориуми", Name: "Име", Size: "Големина", LastModified: "Последна измена"},                  // Macedonian
	"ml":  {ListingName: "ഡയറക്ടറി ലിസ്റ്റിംഗ്", Name: "പേര്", Size: "വലിപ്പം", LastModified: "അവസാനം പുതുക്കിയത്"},                 // Malayalam
	"mn":  {ListingName: "Сангийн жагсаалт", Name: "Нэр", Size: "Хэмжээ", LastModified: "Сүүлд өөрчлөгдсөн"},                        // Mongolian
	"mr":  {ListingName: "निर्देशिका सूची", Name: "नाव", Size: "आकार", LastModified: "शेवटचे बदल"},                                  // Marathi
	"ms":  {ListingName: "Senarai direktori", Name: "Nama", Size: "Saiz", LastModified: "Terakhir diubah"},                          // Malay
	"mt":  {ListingName: "Lista tad-direttorju", Name: "Isem", Size: "Daqs", LastModified: "L-aħħar modifikat"},                     // Maltese
	"my":  {ListingName: "ဖိုင်တွဲစာရင်း", Name: "အမည်", Size: "အရွယ်အစား", LastModified: "နောက်ဆုံးပြင်ဆင်မှု"},                    // Burmese
	"ne":  {ListingName: "डाइरेक्टरी सूची", Name: "नाम", Size: "साइज", LastModified: "अन्तिम परिमार्जन"},                            // Nepali
	"nl":  {ListingName: "Mappinhoud", Name: "Naam", Size: "Grootte", LastModified: "Laatst gewijzigd"},                             // Dutch
	"no":  {ListingName: "Katalogliste", Name: "Navn", Size: "Størrelse", LastModified: "Sist endret"},                              // Norwegian
	"non": {ListingName: "ᛋᚴᚱᛅ", Name: "ᚾᛅᚠᚾ", Size: "ᛋᛏᛅᚱᚦ", LastModified: "ᛒᚱᚢᛏ"},                                                 // Old Norse
	"oc":  {ListingName: "Lista del repertòri", Name: "Nom", Size: "Talha", LastModified: "Darrièra modificacion"},                  // Occitan
	"om":  {ListingName: "Tarree galmee", Name: "Maqaa", Size: "Hamma", LastModified: "Dhuma irratti kan foyya'e"},                  // Oromo
	"or":  {ListingName: "ଡିରେକ୍ଟୋରୀ ତାଲିକା", Name: "ନାମ", Size: "ଆକାର", LastModified: "ଶେଷରେ ପରିବର୍ତ୍ତିତ"},                         // Odia
	"pa":  {ListingName: "ਡਾਇਰੈਕਟਰੀ ਸੂਚੀ", Name: "ਨਾਮ", Size: "ਆਕਾਰ", LastModified: "ਆਖਰੀ ਬਦਲਾਅ"},                                   // Punjabi
	"phn": {ListingName: "𐤔𐤓𐤔 𐤎𐤐𐤓", Name: "𐤔𐤌", Size: "𐤌𐤃𐤃", LastModified: "𐤏𐤕 𐤀𐤇𐤓𐤍", RTL: true},                                    // Phoenician
	"pl":  {ListingName: "Lista katalogu", Name: "Nazwa", Size: "Rozmiar", LastModified: "Ostatnia modyfikacja"},                    // Polish
	"ps":  {ListingName: "د لارښود لړلیک", Name: "نوم", Size: "کچه", LastModified: "وروستی بدلون", RTL: true},                       // Pashto
	"pt":  {ListingName: "Listagem do diretório", Name: "Nome", Size: "Tamanho", LastModified: "Última modificação"},                // Portuguese
	"quc": {ListingName: "Cholaj rech uxe’al", Name: "B’i’aj", Size: "Nimal", LastModified: "Uk’exik k’isbal mul"},                  // Kʼicheʼ
	"quz": {ListingName: "Ñiqichasqa", Name: "Suti", Size: "Sayay", LastModified: "Qhipa llamk'apusqa"},                             // Cusco Quechua
	"ro":  {ListingName: "Listarea directorului", Name: "Nume", Size: "Mărime", LastModified: "Ultima modificare"},                  // Romanian
	"rw":  {ListingName: "Urutonde", Name: "Izina", Size: "Ingano", LastModified: "Iheruka guhindurwa"},                             // Kinyarwanda
	"ru":  {ListingName: "Содержимое директории", Name: "Имя", Size: "Размер", LastModified: "Последнее изменение"},                 // Russian
	"sa":  {ListingName: "निर्देशिकासूची", Name: "नाम", Size: "परिमाणम्", LastModified: "अन्तिमं परिवर्तनम्"},                       // Sanskrit
	"sga": {ListingName: "ᚔᚋᚓᚔᚏᚓ", Name: "ᚐᚔᚋ", Size: "ᚋᚓᚈ", LastModified: "ᚐᚈᚐᚏ"},                                                  // Old Irish
	//nolint:staticcheck //the invisible character here is intended
	"si":  {ListingName: "ඩිරෙක්ටරි ලැයිස්තුව", Name: "නම", Size: "ප්‍රමාණය", LastModified: "අවසන් වරට වෙනස් කළේ"},          // Sinhala
	"sk":  {ListingName: "Výpis adresára", Name: "Názov", Size: "Veľkosť", LastModified: "Posledná zmena"},                  // Slovak
	"sl":  {ListingName: "Seznam imenika", Name: "Ime", Size: "Velikost", LastModified: "Zadnja sprememba"},                 // Slovenian
	"sm":  {ListingName: "Lisi o faila", Name: "Igoa", Size: "Tele", LastModified: "Suiga mulimuli"},                        // Samoan
	"sn":  {ListingName: "Ndandanda", Name: "Zita", Size: "Saizi", LastModified: "Chapedzisira kugadziriswa"},               // Shona
	"so":  {ListingName: "Liiska tusaha", Name: "Magaca", Size: "Baaxadda", LastModified: "Wax ka beddelkii ugu dambeeyay"}, // Somali
	"sq":  {ListingName: "Lista e drejtorisë", Name: "Emri", Size: "Madhësia", LastModified: "Modifikimi i fundit"},         // Albanian
	"sr":  {ListingName: "Списак директоријума", Name: "Назив", Size: "Величинa", LastModified: "Последња измена"},          // Serbian
	"st":  {ListingName: "Lethathamo", Name: "Lebitso", Size: "Boholo", LastModified: "E fetotswe la ho qetela"},            // Sesotho
	"su":  {ListingName: "Daptar diréktori", Name: "Nami", Size: "Ukuran", LastModified: "Parobahan pamungkas"},             // Sundanese
	"sux": {ListingName: "𒁾𒂍 𒈬", Name: "𒈬", Size: "𒃲", LastModified: "𒌓"},                                                   // Sumerian
	"sv":  {ListingName: "Kataloglista", Name: "Namn", Size: "Storlek", LastModified: "Senast ändrad"},                      // Swedish
	"sw":  {ListingName: "Orodha", Name: "Jina", Size: "Ukubwa", LastModified: "Ilibadilishwa mwisho"},                      // Swahili
	"ta":  {ListingName: "கோப்பகப் பட்டியல்", Name: "பெயர்", Size: "அளவு", LastModified: "கடைசியாக மாற்றப்பட்டது"},          // Tamil
	"te":  {ListingName: "డైరెక్టరీ జాబితా", Name: "పేరు", Size: "పరిమాణం", LastModified: "చివరిగా మార్చబడింది"},            // Telugu
	"tg":  {ListingName: "Рӯйхати феҳрист", Name: "Ном", Size: "Андоза", LastModified: "Тағироти охирин"},                   // Tajik
	"th":  {ListingName: "รายการไดเรกทอรี", Name: "ชื่อ", Size: "ขนาด", LastModified: "แก้ไขล่าสุด"},                        // Thai
	"tl":  {ListingName: "Listahan ng direktoryo", Name: "Pangalan", Size: "Laki", LastModified: "Huling binago"},           // Tagalog
	"tk":  {ListingName: "Katalog sanawy", Name: "Ady", Size: "Ölçegi", LastModified: "Soňky üýtgeşme"},                     // Turkmen
	"tr":  {ListingName: "Dizin İçeriği", Name: "Ad", Size: "Boyut", LastModified: "Son değişiklik"},                        // Turkish
	"tt":  {ListingName: "Каталог исемлеге", Name: "Исем", Size: "Зурлык", LastModified: "Соңгы үзгәртү"},                   // Tatar
	"uk":  {ListingName: "Список каталогу", Name: "Ім'я", Size: "Розмір", LastModified: "Остання зміна"},                    // Ukrainian
	"ur":  {ListingName: "ڈائریکٹری فہرست", Name: "نام", Size: "سائز", LastModified: "آخری تبدیلی", RTL: true},              // Urdu
	"uz":  {ListingName: "Direktoriyalar ro'yxati", Name: "Nomi", Size: "Hajmi", LastModified: "Oxirgi o'zgartirish"},       // Uzbek
	"vi":  {ListingName: "Danh sách thư mục", Name: "Tên", Size: "Kích thước", LastModified: "Sửa đổi lần cuối"},            // Vietnamese
	"wo":  {ListingName: "Njataayu", Name: "Tur", Size: "Dayo", LastModified: "Muddit soppi"},                               // Wolof
	"xh":  {ListingName: "Uluhlu", Name: "Igama", Size: "Ubungakanani", LastModified: "Igcinwe ukulungiswa"},                // Xhosa
	"yi":  {ListingName: "דירעקטארי ליסטע", Name: "נאמען", Size: "גרייס", LastModified: "לעצט גענדערט", RTL: true},          // Yiddish
	"yo":  {ListingName: "Àtòjọ ìtọ́kasí", Name: "Oruko", Size: "Iwon", LastModified: "Atunse to kẹhin"},                    // Yoruba
	"zh":  {ListingName: "目录索引", Name: "名称", Size: "大小", LastModified: "修改日期"},                                              // Chinese
	"zu":  {ListingName: "Uhlu", Name: "Igama", Size: "Ubukhulu", LastModified: "Igcine ukulungiswa"},                       // Zulu
}
