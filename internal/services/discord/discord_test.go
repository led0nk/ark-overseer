package discord

//TODO: setup mock server for discord

//func TestDiscordMessages(t *testing.T) {
//	ctx := context.Background()
//
//	tests := []struct {
//		name      string
//		token     string
//		channelID string
//		expectErr bool
//	}{
//		{
//			name:      "valid token and channelID",
//			token:     "dummy-token",
//			channelID: "dummy-channelID",
//			expectErr: false,
//		},
//		{
//			name:      "invalid token",
//			token:     "invalid-token",
//			channelID: "dummy-channelID",
//			expectErr: true,
//		},
//		{
//			name:      "invalid channelID",
//			token:     "dummy-token",
//			channelID: "invalid-channelID",
//			expectErr: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			testDC, err := NewDiscordNotifier(ctx, tt.token, tt.channelID)
//			if tt.expectErr {
//				assert.Error(t, err)
//			}
//			if testDC != nil {
//				err = testDC.Connect(ctx)
//				if tt.expectErr {
//					assert.Error(t, err)
//				}
//			}
//			assert.NoError(t, err)
//		})
//	}
//}
