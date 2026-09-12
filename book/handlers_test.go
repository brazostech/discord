package book

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/brazostech/discord/discord"
	"github.com/brazostech/discord/discord/interactions"
)

func TestHandleRegisterStoresCurrentBook(t *testing.T) {
	store := NewMemoryStore()
	commands := NewCommands(NewService(store))

	packet := commandPacket("server-1", "register",
		stringValueOption("name", "Dune"),
		stringValueOption("url", "https://example.com/dune"),
	)

	res, err := commands.HandleRegister(t.Context(), packet)
	if err != nil {
		t.Fatalf("handle register: %v", err)
	}
	if content := responseContent(t, res); !strings.Contains(content, "Dune") {
		t.Fatalf("response %q does not mention the book", content)
	}

	stored, err := store.CurrentBook(t.Context(), "server-1")
	if err != nil {
		t.Fatalf("current book: %v", err)
	}

	want := Book{Name: "Dune", Url: "https://example.com/dune", Chapter: 0}
	if stored != want {
		t.Fatalf("stored %+v, want %+v", stored, want)
	}
}

func TestHandleRegisterWithoutUrl(t *testing.T) {
	store := NewMemoryStore()
	commands := NewCommands(NewService(store))

	packet := commandPacket("server-1", "register", stringValueOption("name", "Dune"))

	if _, err := commands.HandleRegister(t.Context(), packet); err != nil {
		t.Fatalf("handle register: %v", err)
	}

	stored, err := store.CurrentBook(t.Context(), "server-1")
	if err != nil {
		t.Fatalf("current book: %v", err)
	}
	if stored.Url != "" {
		t.Fatalf("url = %q, want empty", stored.Url)
	}
}

func TestHandleRegisterRequiresName(t *testing.T) {
	store := NewMemoryStore()
	commands := NewCommands(NewService(store))

	packet := commandPacket("server-1", "register")

	res, err := commands.HandleRegister(t.Context(), packet)
	if err != nil {
		t.Fatalf("handle register: %v", err)
	}
	if content := responseContent(t, res); !strings.Contains(content, "required") {
		t.Fatalf("response %q does not explain the missing name", content)
	}

	if _, err := store.CurrentBook(t.Context(), "server-1"); err == nil {
		t.Fatal("a book was stored despite the missing name")
	}
}

func TestHandleUpdateChapterUpdatesCurrentBook(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	commands := NewCommands(service)

	if _, err := service.Register(t.Context(), "server-1", "Dune", ""); err != nil {
		t.Fatalf("register: %v", err)
	}

	packet := commandPacket("server-1", "update-chapter", intValueOption("chapter", 7))

	res, err := commands.HandleUpdateChapter(t.Context(), packet)
	if err != nil {
		t.Fatalf("handle update chapter: %v", err)
	}
	if content := responseContent(t, res); !strings.Contains(content, "7") {
		t.Fatalf("response %q does not mention the chapter", content)
	}

	stored, err := store.CurrentBook(t.Context(), "server-1")
	if err != nil {
		t.Fatalf("current book: %v", err)
	}
	if stored.Chapter != 7 {
		t.Fatalf("chapter = %d, want 7", stored.Chapter)
	}
}

func TestHandleUpdateChapterWithoutCurrentBook(t *testing.T) {
	commands := NewCommands(NewService(NewMemoryStore()))

	packet := commandPacket("server-1", "update-chapter", intValueOption("chapter", 3))

	res, err := commands.HandleUpdateChapter(t.Context(), packet)
	if err != nil {
		t.Fatalf("handle update chapter: %v", err)
	}
	if content := responseContent(t, res); !strings.Contains(content, "No book registered") {
		t.Fatalf("response %q does not explain the missing book", content)
	}
}

func TestHandleUpdateChapterRejectsNegative(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	commands := NewCommands(service)

	if _, err := service.Register(t.Context(), "server-1", "Dune", ""); err != nil {
		t.Fatalf("register: %v", err)
	}

	packet := commandPacket("server-1", "update-chapter", intValueOption("chapter", -1))

	res, err := commands.HandleUpdateChapter(t.Context(), packet)
	if err != nil {
		t.Fatalf("handle update chapter: %v", err)
	}
	if content := responseContent(t, res); !strings.Contains(content, "0 or greater") {
		t.Fatalf("response %q does not explain the invalid chapter", content)
	}

	stored, err := store.CurrentBook(t.Context(), "server-1")
	if err != nil {
		t.Fatalf("current book: %v", err)
	}
	if stored.Chapter != 0 {
		t.Fatalf("chapter = %d, want 0 left unchanged", stored.Chapter)
	}
}

func TestHandleFailsWithoutCommandData(t *testing.T) {
	commands := NewCommands(NewService(NewMemoryStore()))

	packet := interactions.InteractionPacket{}

	if _, err := commands.HandleRegister(t.Context(), packet); err == nil {
		t.Fatal("expected an error for a packet without command data")
	}
	if _, err := commands.HandleUpdateChapter(t.Context(), packet); err == nil {
		t.Fatal("expected an error for a packet without command data")
	}
}

func commandPacket(serverID, subcommand string, opts ...discord.InteractionOption) interactions.InteractionPacket {
	return interactions.InteractionPacket{
		Interaction: interactions.Interaction{
			Type:    interactions.ApplicationCommandInteractionType,
			GuildID: serverID,
		},
		Data: &discord.ApplicationCommandData{
			Name: Command.Name,
			Options: []discord.InteractionOption{
				{
					Name:    subcommand,
					Type:    discord.SubCommandOptionType,
					Options: opts,
				},
			},
		},
	}
}

func stringValueOption(name, value string) discord.InteractionOption {
	raw, _ := json.Marshal(value)

	return discord.InteractionOption{Name: name, Type: discord.StringOptionType, Value: raw}
}

func intValueOption(name string, value int) discord.InteractionOption {
	raw, _ := json.Marshal(value)

	return discord.InteractionOption{Name: name, Type: discord.IntegerOptionType, Value: raw}
}

func responseContent(t *testing.T, res interactions.InteractionResponse) string {
	t.Helper()

	if res.Data == nil {
		t.Fatal("response has no data")
	}

	return res.Data.Content
}
