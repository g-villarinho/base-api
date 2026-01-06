package notification

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"time"
)

type EmailClient interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

//go:embed templates/*
var templateFS embed.FS

type EmailNotification interface {
	SendWelcomeEmail(ctx context.Context, params SendWelcomeEmailParams) error
	SendVerifyEmail(ctx context.Context, params SendVerifyEmailParams) error
	SendResetPasswordEmail(ctx context.Context, params SendResetPasswordEmailParams) error
	SendChangeEmailNotification(ctx context.Context, params SendChangeEmailParams) error
	SendMagicLinkEmail(ctx context.Context, params SendMagicLinkEmailParams) error
}

type emailNotification struct {
	emailClient EmailClient
}

func NewEmailNotification(emailClient EmailClient) EmailNotification {
	return &emailNotification{
		emailClient: emailClient,
	}
}

func (e *emailNotification) SendWelcomeEmail(ctx context.Context, params SendWelcomeEmailParams) error {
	htmlTemplate, err := template.ParseFS(templateFS, "templates/welcome.html")
	if err != nil {
		return fmt.Errorf("parsefs template: %w", err)
	}

	data := struct {
		UserName         string
		RegistrationDate string
		ConfirmationLink string
		CurrentYear      string
		UserEmail        string
	}{
		UserName:         params.Name,
		RegistrationDate: params.UserRegistration.Format("02/01/2006 às 15:04"),
		ConfirmationLink: params.VerificationLink,
		CurrentYear:      time.Now().Format("2006"),
		UserEmail:        params.Email,
	}

	var htmlBuffer bytes.Buffer
	if err := htmlTemplate.Execute(&htmlBuffer, data); err != nil {
		return fmt.Errorf("execute data to html buffer: %w", err)
	}

	subject := "Welcome to voxel! Please, verify your email"

	if err := e.emailClient.SendEmail(ctx, params.Email, subject, htmlBuffer.String()); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

func (e *emailNotification) SendVerifyEmail(ctx context.Context, params SendVerifyEmailParams) error {
	htmlTemplate, err := template.ParseFS(templateFS, "templates/verify-email.html")
	if err != nil {
		return err
	}

	data := struct {
		UserName         string
		ConfirmationLink string
		CurrentYear      string
		UserEmail        string
	}{
		UserName:         params.UserName,
		ConfirmationLink: params.VerificationLink,
		CurrentYear:      time.Now().Format("2006"),
		UserEmail:        params.UserEmail,
	}

	var htmlBuffer bytes.Buffer
	if err := htmlTemplate.Execute(&htmlBuffer, data); err != nil {
		return err
	}

	subject := "Verify your email address"

	if err := e.emailClient.SendEmail(ctx, params.UserEmail, subject, htmlBuffer.String()); err != nil {
		return err
	}

	return nil
}

func (e *emailNotification) SendResetPasswordEmail(ctx context.Context, params SendResetPasswordEmailParams) error {
	htmlTemplate, err := template.ParseFS(templateFS, "templates/reset-password.html")
	if err != nil {
		return err
	}

	data := struct {
		UserName    string
		ResetLink   string
		CurrentYear string
		UserEmail   string
	}{
		UserName:    params.UserName,
		ResetLink:   params.ResetLink,
		CurrentYear: time.Now().Format("2006"),
		UserEmail:   params.UserEmail,
	}

	var htmlBuffer bytes.Buffer
	if err := htmlTemplate.Execute(&htmlBuffer, data); err != nil {
		return err
	}

	subject := "Password Reset Request"

	if err := e.emailClient.SendEmail(ctx, params.UserEmail, subject, htmlBuffer.String()); err != nil {
		return err
	}

	return nil
}

func (e *emailNotification) SendChangeEmailNotification(ctx context.Context, params SendChangeEmailParams) error {
	htmlTemplate, err := template.ParseFS(templateFS, "templates/change-email.html")
	if err != nil {
		return err
	}

	data := struct {
		UserName         string
		NewEmail         string
		ConfirmationLink string
		CurrentYear      string
		UserEmail        string
	}{
		UserName:         params.UserName,
		NewEmail:         params.NewEmail,
		ConfirmationLink: params.ConfirmationLink,
		CurrentYear:      time.Now().Format("2006"),
		UserEmail:        params.UserEmail,
	}

	var htmlBuffer bytes.Buffer
	if err := htmlTemplate.Execute(&htmlBuffer, data); err != nil {
		return err
	}

	subject := "Email Change Confirmation"

	if err := e.emailClient.SendEmail(ctx, params.UserEmail, subject, htmlBuffer.String()); err != nil {
		return err
	}

	return nil
}

func (e *emailNotification) SendMagicLinkEmail(ctx context.Context, params SendMagicLinkEmailParams) error {
	htmlTemplate, err := template.ParseFS(templateFS, "templates/magic-link.html")
	if err != nil {
		return fmt.Errorf("parse magic-link template: %w", err)
	}

	data := struct {
		UserName    string
		LoginLink   string
		CurrentYear string
		UserEmail   string
	}{
		UserName:    params.UserName,
		LoginLink:   params.LoginLink,
		CurrentYear: time.Now().Format("2006"),
		UserEmail:   params.UserEmail,
	}

	var htmlBuffer bytes.Buffer
	if err := htmlTemplate.Execute(&htmlBuffer, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	subject := "Your login link"

	if err := e.emailClient.SendEmail(ctx, params.UserEmail, subject, htmlBuffer.String()); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}
