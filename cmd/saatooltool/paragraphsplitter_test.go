package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParagraphSplitter_Split(t *testing.T) {
	ps := NewParagraphSplitter()
	text := "one\ntwo\nthree\nfour.\n\nfive\nsix\nseven\nine."
	paragraphs := ps.Split(text)
	assert.Len(t, paragraphs, 2)

	text = `This is a very long text. It should be split into multiple paragraphs based on the maximum number of words allowed per paragraph. Each paragraph should not exceed the specified limit. The number of words in the example is 103 words. It contains a list of cars: Ford, Toyota, Honda, BMW, Audi, Mercedes, Volkswagen, Chevrolet, Nissan, Hyundai. A list of countries: USA, Canada, Mexico, Brazil, Argentina, UK, France, Germany, Italy, Spain. A list of fruits: Apple, Banana, Orange, Grape, Mango, Pineapple, Strawberry, Blueberry, Raspberry, Blackberry. A list of vegetables: Carrot, Broccoli, Spinach, Kale, Lettuce, Cabbage, Cauliflower, Peas, Corn, Potatoes. A list of animals: Dog, Cat, Elephant, Tiger, Lion, Bear, Wolf, Fox, Deer, Rabbit. A list of colors: Red, Blue, Green, Yellow, Orange, Purple, Pink, Brown, Black, White. A list of shapes: Circle, Square, Triangle, Rectangle, Oval, Diamond, Star, Heart, Crescent, Hexagon. A list of professions: Doctor, Engineer, Teacher, Lawyer, Artist, Musician, Writer, Scientist, Chef, Pilot. A list of sports: Soccer, Basketball, Baseball, Tennis, Golf, Swimming, Running, Cycling, Volleyball, Football. A list of hobbies: Reading, Traveling, Cooking, Gardening, Painting, Dancing, Singing, Hiking, Fishing, Photography. A list of musical instruments: Guitar, Piano, Drums, Violin, Flute, Saxophone, Trumpet, Cello, Clarinet, Harp. A list of programming languages: Go, Python, Java, C++, JavaScript, Ruby, Swift, Kotlin, PHP, Rust. A list of operating systems: Windows, macOS, Linux, Android, iOS, Ubuntu, Fedora, Debian, CentOS, Red Hat. A list of web browsers: Chrome, Firefox, Safari, Edge, Opera, Brave, Vivaldi, Tor, Internet Explorer, Netscape. A list of databases: MySQL, PostgreSQL, MongoDB, SQLite, Oracle, SQL Server, Redis, Cassandra, MariaDB, CouchDB. A list of cloud providers: AWS, Azure, Google Cloud, IBM Cloud, Oracle Cloud, DigitalOcean, Linode, Vultr, Heroku, Firebase. A list of social media platforms: Facebook, Twitter, Instagram, LinkedIn, TikTok, Snapchat, Pinterest, Reddit, Tumblr, WhatsApp. A list of email providers: Gmail, Outlook, Yahoo Mail, ProtonMail, Zoho Mail, AOL Mail, GMX Mail, Yandex Mail, Mail.com, iCloud Mail. A list of messaging apps: WhatsApp, Messenger, Telegram, Signal, WeChat, Viber, Line, Kik, Discord, Slack. A list of video conferencing tools: Zoom, Microsoft Teams, Google Meet, Skype, Webex, GoToMeeting.`

	paragraphs = ps.Split(text)
	// assert.Len(t, paragraphs, 10)
	for _, p := range paragraphs {
		fmt.Printf("Paragraph (%d words): %s\n", len(strings.Fields(p)), p)
	}
}
