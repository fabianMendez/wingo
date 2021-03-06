package email

const TplPreffix = `
<!doctype html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">

<head>
  <title>
  </title>
  <!--[if !mso]><!-->
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <!--<![endif]-->
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style type="text/css">
    #outlook a {
      padding: 0;
    }

    body {
      margin: 0;
      padding: 0;
      -webkit-text-size-adjust: 100%;
      -ms-text-size-adjust: 100%;
    }

    table,
    td {
      border-collapse: collapse;
      mso-table-lspace: 0pt;
      mso-table-rspace: 0pt;
    }

    img {
      border: 0;
      height: auto;
      line-height: 100%;
      outline: none;
      text-decoration: none;
      -ms-interpolation-mode: bicubic;
    }

    p {
      display: block;
      margin: 13px 0;
    }
  </style>
  <!--[if mso]>
        <noscript>
        <xml>
        <o:OfficeDocumentSettings>
          <o:AllowPNG/>
          <o:PixelsPerInch>96</o:PixelsPerInch>
        </o:OfficeDocumentSettings>
        </xml>
        </noscript>
        <![endif]-->
  <!--[if lte mso 11]>
        <style type="text/css">
          .mj-outlook-group-fix { width:100% !important; }
        </style>
        <![endif]-->
  <!--[if !mso]><!-->
  <link href="https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700" rel="stylesheet" type="text/css">
  <style type="text/css">
    @import url(https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700);
  </style>
  <!--<![endif]-->
  <style type="text/css">
    @media only screen and (min-width:480px) {
      .mj-column-per-100 {
        width: 100% !important;
        max-width: 100%;
      }
    }
  </style>
  <style media="screen and (min-width:480px)">
    .moz-text-html .mj-column-per-100 {
      width: 100% !important;
      max-width: 100%;
    }
  </style>
  <style type="text/css">
  </style>
</head>

<body style="word-spacing:normal;">
  <div style="">
    <!--[if mso | IE]><table align="center" border="0" cellpadding="0" cellspacing="0" class="" style="width:600px;" width="600" bgcolor="#FAFAFA" ><tr><td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;"><![endif]-->
    <div style="background:#FAFAFA;background-color:#FAFAFA;margin:0px auto;max-width:600px;">
      <table align="center" border="0" cellpadding="0" cellspacing="0" role="presentation" style="background:#FAFAFA;background-color:#FAFAFA;width:100%;">
        <tbody>
          <tr>
            <td style="direction:ltr;font-size:0px;padding:20px 0;text-align:center;">
              <!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="" style="vertical-align:top;width:600px;" ><![endif]-->
              <div class="mj-column-per-100 mj-outlook-group-fix" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;">
                <table border="0" cellpadding="0" cellspacing="0" role="presentation" style="vertical-align:top;" width="100%">
`

const TplCancelSubscription = `
<tr>
  <td align="center" style="font-size:0px;padding:10px 25px;word-break:break-word;">
    <p style="border-top:dashed 1px lightgrey;font-size:1px;margin:0px auto;width:100%;">
    </p>
    <!--[if mso | IE]><table align="center" border="0" cellpadding="0" cellspacing="0" style="border-top:dashed 1px lightgrey;font-size:1px;margin:0px auto;width:550px;" role="presentation" width="550px" ><tr><td style="height:0;line-height:0;"> &nbsp;
</td></tr></table><![endif]-->
  </td>
</tr>
<tr>
  <td align="center" vertical-align="middle" style="font-size:0px;padding:10px 25px;word-break:break-word;">
    <table border="0" cellpadding="0" cellspacing="0" role="presentation" style="border-collapse:separate;line-height:100%;">
      <tr>
        <td align="center" bgcolor="#D1D5DB" role="presentation" style="border:none;border-radius:3px;cursor:auto;mso-padding-alt:10px 25px;background:#D1D5DB;" valign="middle">
          <a href="{{.CancelSubscriptionLink}}" style="display:inline-block;background:#D1D5DB;color:#134E4A;font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:13px;font-weight:normal;line-height:120%;margin:0;text-decoration:none;text-transform:none;padding:10px 25px;mso-padding-alt:0px;border-radius:3px;" target="_blank"> Cancelar suscripci??n </a>
        </td>
      </tr>
    </table>
  </td>
</tr>
`

const TplSuffix = `
                </table>
              </div>
              <!--[if mso | IE]></td></tr></table><![endif]-->
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <!--[if mso | IE]></td></tr></table><![endif]-->
  </div>
</body>

</html>
`
const TplPriceChange = TplPreffix + `
<tbody>
<tr>
  <td align="left" style="font-size:0px;padding:10px 25px;word-break:break-word;">
	<div style="font-family:Helvetica;font-size:18px;font-weight:bold;line-height:1;text-align:left;color:#4B5563;">{{.Message}}</div>
  </td>
</tr>
<tr>
  <td align="center" vertical-align="middle" style="font-size:0px;padding:10px 25px;word-break:break-word;">
	<table border="0" cellpadding="0" cellspacing="0" role="presentation" style="border-collapse:separate;line-height:100%;">
	  <tr>
      <td align="center" bgcolor="#4068E0" role="presentation" style="border:none;border-radius:3px;cursor:auto;mso-padding-alt:10px 25px;background:#4068E0;" valign="middle">
        <a href="{{.LinkHistory}}" style="display:inline-block;background:#4068E0;color:#ffffff;font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:13px;font-weight:bold;line-height:120%;margin:0;text-decoration:none;text-transform:none;padding:10px 25px;mso-padding-alt:0px;border-radius:3px;" target="_blank">Historial</a>
      </td>
    </tr>
  </table>
  </td>
</tr>
<tr>
  <td align="center" vertical-align="middle" style="font-size:0px;padding:10px 25px;word-break:break-word;">
	<table border="0" cellpadding="0" cellspacing="0" role="presentation" style="border-collapse:separate;line-height:100%;">
	  <tr>
		<td align="center" bgcolor="#14B8A6" role="presentation" style="border:none;border-radius:3px;cursor:auto;mso-padding-alt:10px 25px;background:#14B8A6;" valign="middle">
		  <a href="{{.Link}}" style="display:inline-block;background:#14B8A6;color:#ffffff;font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:13px;font-weight:bold;line-height:120%;margin:0;text-decoration:none;text-transform:none;padding:10px 25px;mso-padding-alt:0px;border-radius:3px;" target="_blank">Ver en Wingo</a>
		</td>
	  </tr>
	</table>
  </td>
</tr>
</tbody>
` + TplCancelSubscription + TplSuffix

const TplConfirmSubscription = TplPreffix + `
<tbody>
  <tr>
    <td align="left" style="font-size:0px;padding:10px 25px;word-break:break-word;">
      <div style="font-family:Helvetica;font-size:26px;font-weight:bolder;line-height:1;text-align:left;color:#111827;">Confirma tu suscripci??n</div>
    </td>
  </tr>
  <tr>
    <td align="left" style="font-size:0px;padding:10px 25px;word-break:break-word;">
      <div style="font-family:Helvetica;font-size:18px;line-height:1;text-align:left;color:#4B5563;">
      Usa el siguiente link para confirmar tu suscripci??n para recibir notificaciones sobre actualizaciones del precio de la ruta {{.subscription.Origin}} -> {{.subscription.Destination}} el {{.subscription.Date}}:
      </div>
    </td>
  </tr>
  <tr>
    <td align="center" vertical-align="middle" style="font-size:0px;padding:10px 25px;word-break:break-word;">
      <table border="0" cellpadding="0" cellspacing="0" role="presentation" style="border-collapse:separate;line-height:100%;">
        <tr>
          <td align="center" bgcolor="#14B8A6" role="presentation" style="border:none;border-radius:3px;cursor:auto;mso-padding-alt:10px 25px;background:#14B8A6;" valign="middle">
            <a href="{{.link}}" style="display:inline-block;background:#14B8A6;color:#ffffff;font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:13px;font-weight:bold;line-height:120%;margin:0;text-decoration:none;text-transform:none;padding:10px 25px;mso-padding-alt:0px;border-radius:3px;" target="_blank">
            Confirmar
            </a>
          </td>
        </tr>
      </table>
    </td>
  </tr>
  <tr>
    <td align="center" style="font-size:0px;padding:10px 25px;word-break:break-word;">
      <p style="border-top:dashed 1px lightgrey;font-size:1px;margin:0px auto;width:100%;">
      </p>
      <!--[if mso | IE]><table align="center" border="0" cellpadding="0" cellspacing="0" style="border-top:dashed 1px lightgrey;font-size:1px;margin:0px auto;width:550px;" role="presentation" width="550px" ><tr><td style="height:0;line-height:0;"> &nbsp;
</td></tr></table><![endif]-->
    </td>
  </tr>
  <tr>
    <td align="left" style="font-size:0px;padding:10px 25px;word-break:break-word;">
      <div style="font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:13px;line-height:1;text-align:left;color:#4B5563;">
        Si no solicitaste esta suscripci??n, por favor ignora este mensaje.
      </div>
    </td>
  </tr>
</tbody>
` + TplSuffix

const TplConfirmSubscriptionEn = TplPreffix + `
<tbody>
  <tr>
    <td align="left" style="font-size:0px;padding:10px 25px;word-break:break-word;">
      <div style="font-family:Helvetica;font-size:26px;font-weight:bolder;line-height:1;text-align:left;color:#111827;">Confirm your subscription</div>
    </td>
  </tr>
  <tr>
    <td align="left" style="font-size:0px;padding:10px 25px;word-break:break-word;">
      <div style="font-family:Helvetica;font-size:18px;line-height:1;text-align:left;color:#4B5563;">Use the following link to confirm your subscription to receive notifications about price updates for the route {{.subscription.Origin}} -> {{.subscription.Destination}} on {{.subscription.Date}}:</div>
    </td>
  </tr>
  <tr>
    <td align="center" vertical-align="middle" style="font-size:0px;padding:10px 25px;word-break:break-word;">
      <table border="0" cellpadding="0" cellspacing="0" role="presentation" style="border-collapse:separate;line-height:100%;">
        <tr>
          <td align="center" bgcolor="#14B8A6" role="presentation" style="border:none;border-radius:3px;cursor:auto;mso-padding-alt:10px 25px;background:#14B8A6;" valign="middle">
            <a href="{{.link}}" style="display:inline-block;background:#14B8A6;color:#ffffff;font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:13px;font-weight:bold;line-height:120%;margin:0;text-decoration:none;text-transform:none;padding:10px 25px;mso-padding-alt:0px;border-radius:3px;" target="_blank"> Confirm </a>
          </td>
        </tr>
      </table>
    </td>
  </tr>
  <tr>
    <td align="center" style="font-size:0px;padding:10px 25px;word-break:break-word;">
      <p style="border-top:dashed 1px lightgrey;font-size:1px;margin:0px auto;width:100%;">
      </p>
      <!--[if mso | IE]><table align="center" border="0" cellpadding="0" cellspacing="0" style="border-top:dashed 1px lightgrey;font-size:1px;margin:0px auto;width:550px;" role="presentation" width="550px" ><tr><td style="height:0;line-height:0;"> &nbsp;
</td></tr></table><![endif]-->
    </td>
  </tr>
  <tr>
    <td align="left" style="font-size:0px;padding:10px 25px;word-break:break-word;">
      <div style="font-family:Ubuntu, Helvetica, Arial, sans-serif;font-size:13px;line-height:1;text-align:left;color:#4B5563;">If you did not request this subscription, please ignore this message.</div>
    </td>
  </tr>
</tbody>
` + TplSuffix
