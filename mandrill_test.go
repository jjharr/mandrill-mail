package mandrillmail

import (
	"errors"
	"html/template"
	"net/http"
	"testing"
)

// You must set these parameters for your own testing
const (
	mandrillTestKey   = ``
	testDomain        = ``
	testFromEmail     = ``
	testToFirstEmail  = ``
	testToSecondEmail = ``
	testToThirdEmail  = ``
	testTemplate      = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
    <meta name="viewport" content="width=device-width" />
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <title>Alerts e.g. approaching your limit</title>
    <link href="styles.css" media="all" rel="stylesheet" type="text/css" />
</head>

<body itemscope itemtype="http://schema.org/EmailMessage">

<table class="body-wrap">
    <tr>
        <td></td>
        <td class="container" width="600">
            <div class="content">
                <table class="main" width="100%" cellpadding="0" cellspacing="0">
                    <tr>
                        <td class="alert alert-warning">
                            Alert: This test is {.TestParam}.
                        </td>
                    </tr>
                    <tr>
                        <td class="content-wrap">
                            <table width="100%" cellpadding="0" cellspacing="0">
                                <tr>
                                    <td class="content-block">
                                        You have <strong>1 free report</strong> remaining.
                                    </td>
                                </tr>
                                <tr>
                                    <td class="content-block">
                                        Add your credit card now to upgrade your account to a premium plan to ensure you don't miss out on any reports.
                                    </td>
                                </tr>
                                <tr>
                                    <td class="content-block">
                                        <a href="http://www.mailgun.com" class="btn-primary">Upgrade my account</a>
                                    </td>
                                </tr>
                                <tr>
                                    <td class="content-block">
                                        Thanks for choosing Acme Inc.
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                </table>
                <div class="footer">
                    <table width="100%">
                        <tr>
                            <td class="aligncenter content-block"><a href="http://www.mailgun.com">Unsubscribe</a> from these alerts.</td>
                        </tr>
                    </table>
                </div></div>
        </td>
        <td></td>
    </tr>
</table>

</body>
</html>
`
)

var testRecipients = []MailRecipient{
	{
		Email:         testToFirstEmail,
		Name:          `Test To User`,
		RecipientType: MAIL_TO,
	},
	{
		Email:         testToSecondEmail,
		Name:          `Test CC User`,
		RecipientType: MAIL_CC,
	},
	{
		Email:         testToThirdEmail,
		Name:          `Test BCC User`,
		RecipientType: MAIL_BCC,
	},
}

var testMessage = MailMessage{
	HTMLTemplate: nil,
	TextTemplate: nil,
	TemplateVars: map[string]string{
		`TestParam`: `Working`,
	},
	AutoText: true,
	Subject:  `Mandrill Mail Test`,
	From: &MailRecipient{
		Name:  `Mandrill Name`,
		Email: testFromEmail,
	},
	ReplyTo:       testFromEmail,
	MarkImportant: true,
	Attachments: []EmailAttachment{
		{
			Name:          `Test Attachment`,
			MimeType:      `image/png`,
			Base64Content: `data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAMUAAAAvCAYAAABT5w4UAAAKq2lDQ1BJQ0MgUHJvZmlsZQAASImVlgdUU2kWx7/3XnqhJURASui9S5deQ+9NVEJCCSWEQFCxK4MjOKKIiIA6olIVHJUio4iIYhsELNgHZBBQ18GCDZV9wBJ2ds/unv3n3Nzfufneffd9ed85fwAo99kCQSosBUAaP0sY7OnCjIyKZuKHAARUABkgwIrNyRQ4Bwb6AlTz+a/6cA9djeq24Uyvf//9v0qaG5/JAQAKRDmOm8lJQ/kMGu0cgTALAAQNoL4qSzDDpSjTheiAKB+f4cQ57pjhuDm+M7smNNgV5TEACBQ2W5gIAPk9WmdmcxLRPhQ6yiZ8Lo+PshvKDpwkNhflXJQN0tLSZ/gkyjpx/9Qn8S8948Q92exEMc89y6wIbrxMQSp7zf+5Hf9baami+XuooUFJEnoFo5mB7llNSrqPmPlx/gHzzOPOrp/lJJFX2DxzMl2j55nLdvOZZ1FKmPM8s4UL1/KyWKHzLEwPFvfnp/r7ivvHs8Qcn+keMs8JPA/WPOckhUbMczYv3H+eM1NCfBbWuIrrQlGweOYEoYf4GdMyF2bjsBfulZUU6rUwQ6R4Hm68m7u4zg8TrxdkuYh7ClIDF+ZP9RTXM7NDxNdmoS/YPCezvQMX+gSK9we4AXfgi36YIAyYAStgCiwAOlVW/OqZdxq4pgvWCHmJSVlMZ/TUxDNZfI6RAdPMxNQKgJkzOPcXv7s/e7YgBmGhlt4LgFUtCtULNXYsAK3obsiqL9Q0jwEg+QcA5zkckTB7roaZ+cICEpAEdCAPlIE60AGG6HyWwA44oRN7gwAQCqLACsABSSANCMEqsA5sBnmgAOwCe0EZOASOgBpwApwCLeAcuAiugBugF9wFj8AgGAEvwQT4AKYgCMJDVIgGyUMqkCakD5lB1pAD5A75QsFQFBQLJUJ8SAStg7ZCBVARVAYdhmqhX6Cz0EXoGtQHPYCGoHHoLfQFRmAKTIeVYC3YGLaGnWEfOBReDifCGXAOnAvvhEvhSvg43AxfhG/Ad+FB+CU8iQCEjDAQVcQQsUZckQAkGklAhMgGJB8pQSqRBqQN6UZuI4PIK+QzBoehYZgYQ4wdxgsThuFgMjAbMDswZZgaTDOmC3MbM4SZwHzHUrGKWH2sLZaFjcQmYldh87Al2CpsE/Yy9i52BPsBh8MxcNo4K5wXLgqXjFuL24E7gGvEdeD6cMO4STweL4/Xx9vjA/BsfBY+D78ffxx/Ad+PH8F/IpAJKgQzggchmsAnbCGUEOoI7YR+wihhiihF1CTaEgOIXOIaYiHxKLGNeIs4QpwiSZO0SfakUFIyaTOplNRAukx6THpHJpPVyDbkIDKPvIlcSj5JvkoeIn+myFD0KK6UGIqIspNSTemgPKC8o1KpWlQnajQ1i7qTWku9RH1K/SRBkzCSYElwJTZKlEs0S/RLvJYkSmpKOkuukMyRLJE8LXlL8pUUUUpLylWKLbVBqlzqrNSA1KQ0TdpUOkA6TXqHdJ30NekxGbyMloy7DFcmV+aIzCWZYRpCU6e50ji0rbSjtMu0ETqOrk1n0ZPpBfQT9B76hKyM7BLZcNnVsuWy52UHGQhDi8FipDIKGacY9xhfFiktcl4Uv2j7ooZF/Ys+yi2Wc5KLl8uXa5S7K/dFninvLp8iv1u+Rf6JAkZBTyFIYZXCQYXLCq8W0xfbLeYszl98avFDRVhRTzFYca3iEcWbipNKykqeSgKl/UqXlF4pM5SdlJOVi5XblcdVaCoOKjyVYpULKi+YskxnZiqzlNnFnFBVVPVSFakeVu1RnVLTVgtT26LWqPZEnaRurZ6gXqzeqT6hoaLhp7FOo17joSZR01ozSXOfZrfmRy1trQitbVotWmPactos7Rzteu3HOlQdR50MnUqdO7o4XWvdFN0Dur16sJ6FXpJeud4tfVjfUp+nf0C/zwBrYGPAN6g0GDCkGDobZhvWGw4ZMYx8jbYYtRi9NtYwjjbebdxt/N3EwiTV5KjJI1MZU2/TLaZtpm/N9Mw4ZuVmd8yp5h7mG81bzd8s0V8Sv+TgkvsWNAs/i20WnRbfLK0shZYNluNWGlaxVhVWA9Z060DrHdZXbbA2LjYbbc7ZfLa1tM2yPWX7p52hXYpdnd3YUu2l8UuPLh22V7Nn2x+2H3RgOsQ6/Oww6KjqyHasdHzmpO7EdapyGnXWdU52Pu782sXERejS5PLR1dZ1vWuHG+Lm6Zbv1uMu4x7mXub+1EPNI9Gj3mPC08JzrWeHF9bLx2u31wBLicVh1bImvK2813t3+VB8QnzKfJ756vkKfdv8YD9vvz1+j/01/fn+LQEggBWwJ+BJoHZgRuCvQbigwKDyoOfBpsHrgrtDaCErQ+pCPoS6hBaGPgrTCROFdYZLhseE14Z/jHCLKIoYjDSOXB95I0ohihfVGo2PDo+uip5c5r5s77KRGIuYvJh7y7WXr15+bYXCitQV51dKrmSvPB2LjY2IrYv9yg5gV7In41hxFXETHFfOPs5LrhO3mDsebx9fFD+aYJ9QlDCWaJ+4J3E8yTGpJOkVz5VXxnuT7JV8KPljSkBKdcp0akRqYxohLTbtLF+Gn8LvSldOX53eJ9AX5AkGM2wz9mZMCH2EVZlQ5vLM1iw6anZuinREP4iGsh2yy7M/rQpfdXq19Gr+6ptr9NZsXzOa45FzbC1mLWdt5zrVdZvXDa13Xn94A7QhbkPnRvWNuRtHNnluqtlM2pyy+bctJluKtrzfGrG1LVcpd1Pu8A+eP9TnSeQJ8wa22W079CPmR96PPdvNt+/f/j2fm3+9wKSgpODrDs6O6z+Z/lT60/TOhJ09hZaFB3fhdvF33dvtuLumSLoop2h4j9+e5mJmcX7x+70r914rWVJyaB9pn2jfYKlvaet+jf279n8tSyq7W+5S3lihWLG94uMB7oH+g04HGw4pHSo49OVn3s/3D3sebq7Uqiw5gjuSfeT50fCj3cesj9VWKVQVVH2r5lcP1gTXdNVa1dbWKdYV1sP1ovrx4zHHe0+4nWhtMGw43MhoLDgJTopOvvgl9pd7p3xOdZ62Pt1wRvNMRROtKb8Zal7TPNGS1DLYGtXad9b7bGebXVvTr0a/Vp9TPVd+XvZ8YTupPbd9+kLOhckOQceri4kXhztXdj66FHnpTldQV89ln8tXr3hcudTt3H3hqv3Vc9dsr529bn295YbljeabFjebfrP4ranHsqf5ltWt1l6b3ra+pX3t/Y79F2+73b5yh3Xnxl3/u333wu7dH4gZGLzPvT/2IPXBm4fZD6cebXqMfZz/ROpJyVPFp5W/6/7eOGg5eH7Ibejms5Bnj4Y5wy//yPzj60juc+rzklGV0doxs7Fz4x7jvS+WvRh5KXg59Srvb9J/q3it8/rMn05/3pyInBh5I3wz/XbHO/l31e+XvO+cDJx8+iHtw9TH/E/yn2o+W3/u/hLxZXRq1Vf819Jvut/avvt8fzydNj0tYAvZs1YAQQNOSADgbTUA1CgAaKivIEnMeeRZQXO+fpbAf+I5Hz0rSwBqnQCYsWp+aD6AZi00S6IRiEaoE4DNzcXxD2UmmJvN9SK3oNakZHr6HeoN8boAfBuYnp5qmZ7+VoUO+xCAjg9z3nxGUqj/7403dYv07ac6gH/V3wGxsQTLPAn0qQAAAZxpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IlhNUCBDb3JlIDUuNC4wIj4KICAgPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4KICAgICAgPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIKICAgICAgICAgICAgeG1sbnM6ZXhpZj0iaHR0cDovL25zLmFkb2JlLmNvbS9leGlmLzEuMC8iPgogICAgICAgICA8ZXhpZjpQaXhlbFhEaW1lbnNpb24+MTk3PC9leGlmOlBpeGVsWERpbWVuc2lvbj4KICAgICAgICAgPGV4aWY6UGl4ZWxZRGltZW5zaW9uPjQ3PC9leGlmOlBpeGVsWURpbWVuc2lvbj4KICAgICAgPC9yZGY6RGVzY3JpcHRpb24+CiAgIDwvcmRmOlJERj4KPC94OnhtcG1ldGE+CuOqeMkAAA4wSURBVHgB7VtdaBtXFv6ypCBDHmRowVryUC8p1IHCjtkuxHlpbFKwSspWxQ9r032R3aUkDQS7hazcwqY2BVfZQrC34NgsmyIXGuRCi2xIkfpQrCwJSEtapIUsdiHBMtggQQwaSGD23Jm5M3f+bFm20pq9F+SZuT/nnvudc+7POddHNEqQSSIgEbAQ+JX1Jl8kAhIBHQFpFFIRJAIuBKRRuACRnxIBaRRSByQCLgSkUbgAkZ8SAWkUUgckAi4EpFG4AJGfEgFpFFIHJAIuBKRRuACRnxIBaRRSByQCLgSkUbgAkZ8SAWkUUgckAi4EpFG4AJGfEgFpFFIHJAIuBKRRuACRnxKBo41CUPl2Bh/cXEPHs6FGm0DdUhH9MIne4w03ab7ikxpqahjhY02SeLKGmfcnsRYKo7YVwejfRtHVAK3aRg3hjrC30/3y46XYWM6TCubf/wC3t8MY/HCCsG9cXo118EutpaKyBUT2oJ+BI2H/ZNRIKs3G2D8j7fk3d6/eCPl91VnPp7Qo8aZcKzRNp5pPOsYWX1jdmdajdS11OUptFK3gGuJB8LNz5zuU1ktazJTT08B+B06eWlH9wYqW6Ge6Oae5RNEUDw2vFKGObkT7gZA1e7YD2/NYXLbtLToQRwhVK0PdBtrbrM/WvKhFDPcMYYmox0LNzooqlv4x5uBvfnABH/0xgYgj1/4ofzGMoY/1XmnMQjoQfgR6e30liZJk/o+SitTrpzFZpCEPHIyyNWwUnecSyJxzYf1kFMPPnMQ8y+5PIfPloKvCU/g8GgpU3IZ738ojed1dexxLxVHEFYfK25VCAeZyEPzYvci3BhBoO0GVmFEcUGrYKHz7e+LMVekzQIWAJ7Tn26hAZZUotXd07rr/V2sVVOlcohLVULjdf79IBPnapHLiRhcN/y0vL5iYKkjlk8j19OmGPvx5joyCNmY+SVV5r4Aq4tAoP4RHbauK6jaNjq1woTYan8/ZxN33do1wrBImbNUmTPzOM7zNM4Y01K0KKjVDOu0dkWDc2Ti4Rlj9hEhWrjZMlj+RLFk/oXZ0Ht+Fb5V4fmjwzDQk8nwEId4P55U/RR7oXGb0Q/IPHKuKOu1I9LRdN3DhtJp9NrXp4o0el7Q4P2f0pwL3c4XFpKbwesIzdjmlrftsAtfvpLSY4nd+iWrJxRLvXVtdGHWcAwgD/Tt2w65jVQ58WdcmLJ4mtCrVS4/wvhUtyzLEJI7Zasfqx7Rb/2yAn2pJm77EziK8D/GpOMYndlstZ7XEgOJtp8S11J11u6rA3/StFS11ydtGGUlqpUd2E40klxow+Bi9kdXSU97zY5RkxURVJdl4ZKmMaisPfAT5uKplpuJenmnsiRsrHn0p3TDqKpfTWiE359NO0aZvCWe9+ymfOjSOgWBdFEcd9I6ggobyBQHQ9skzSAZ2+pIodL/3uFYQFK9a8APD2S6+YCh90OF/Lwfu+j0b2NisQVc8dHsMjMbMDvVepY5q6RmvMrF6Fj+bKwFtnfQStwQlJ0HUG8DEaiPKxJdPsy9lWp8ADDnbRuEdl5O34PKEtv5Y0Bo68I/u1D8rI+UVRK9xowjuw+Bl+o7Rql4O0BVlfwfulhrF+qI4c8a1bNmEoL6upa8Is6Vl2aJwYlqmQMrBgKYf8+jYM9SEpqsN5dcfZC1FU65ktXq9rtVF4Qhy8nvNXuYzKa0KFbOGQ7FcwmZVHte17BTnP6plH1S1+iOaKXfhJ3uF9wUtfi2rrfPZmvDIzgpYWXiwvoTVmBQpPpvVqmx8xEPp1rRgnBwTZ330Jwh3A8f1exnLM8UUz/ZOibgzxTOwZzg6cTeUks3yVRpuvVLQJnSvj5Fv09O0jDgZjsxppU1jJXG34RMcg9VtFHo/hGudPH2ZK8KEM5Jm1fVUf7SqJfmuQklqq3p9Xtrcs4VGIW5LYlqBK4DAZ1YwjIyukIJwSDHcqXDNUESlf0IrccUXlMYzq7sJuL/rBVtJXCvdiqX00JLmzCQ2twUYs3lhFYL4EfKBad9VNcUVTOBl/WvbWGKzXpdzaZYbJyl5mRmmaBQJJ2/EnrgKplh9PQm4k4s5fd/MNh8rgjHz1ZTXqN+zZ2vLKDaz9gTWP+dYDYx267Yiw550bEyhjS4K2yS9UdVuI+DDdiN867ffbRMfU+si2htFjNN0xJIyNQ7FcuUaeexvzx/Io26m/A8V440fmm4O4bW3ryJ3t4yamae8m2FGjMJSAl1+BzXzEM9p7vasfJvGolkp+UHU4SToees8aGXS09gnS8ah0vz2PMSDtlgo8nO0C9OPqli/X8Lq5nlHX6D2tY0y1jbExsZ7rVYzM6MYHeAc2fW6BpLIfJ1FobyKoRNON0fsxpAHp7DSC4566b7tLLAoDowhesL60l/an+82MxgPXY7C0AvdFj1eUCnmLGdQ8q9vwnsMjyD6DndgLKFIQTdnimGov9OZRVS6XvGO31GJ644jc+8ffqq1dyo+LdRNLkzylr3fjSNfUKUd3Gbln1j9CM68R8feZcOclq6Pgf1YUvrjiL81iOjZHnQeRNQSNSx9NqnTZn/GbqYQ+TGEupnTpt622b2ZRH5jEL0dVvWmXkLHwoicCKPyQx4LXxWxtraGwr/LFOth8Q6/pKL8ve7wJpd3FCe92gWEuxA9JyhqkIFy8uQybjffyxSNZ5g7EnlwghO1ddodeatserwdM3Cexn7fDkOCPMf9LKKySbNHh0i43TlpmE06X2bGuYMSuUk3+d0yo8AzLo52Gcua6U6NnElg9VYHLrw6rAfkOJXi8jwu0I+l2LUVpN/t4UXNPR/mMCwEHvHpBQwFUiri6s0iet/dZaYKbG8WbOQxHDltxHV2q2uWkzfTSOSSFNWmkebR37pn291bRfu799yPlyo3O2+JX47Dpc0qDPSi02+wuxm8H/Em8lpnFAIz8RsFfHS2Ayr55MWk++c5B0IwrPNsHBly9tYellG4m0dueRGT1+3ZdPHiacy8Usf5l/yQE3sIfi9+Zc7AVEWhSHyvn699ew1XrxsbrKWLKZTfUTzbkeAeXCV0J2nSYRBRJKZi6H6xE53Pd6Krqx2Lr7djSDRUkcS2GZsQ8/g7VxaOJc9v4qnLpIl2zib2tix1ZxVnaDFyh5As2RPvbY5Vgigd0DbIyVPjXwcAo39noWO2wlZototQAMiTSFHy35Wh0izY+QLNLgRQuZhH+T908fDMEHqOd5Gy0u+NOCau1ZD7fAJ9dM5gKfevNTIKYdvgIb5DBl3+S13kRjaK9JdJBM2pPTiCN/Vo91Usfj+OxBm/PcwOfZlF6n+/s85YGEmhOjvo3Wvzcxd/0pzddZZOADfJMGmLVaqdR4+7+60cup/r0zcV5MnB3MDuvLS6RuiYjWaVtmcRv0uJtMXK3a0gdDyCk8+SbrRME/c+2tYdtI+fQsLkZ+niOHKewxQp9sev4fSrfejr6cbMbZpdtvM4+fs+vPmnYZz+e945Grq92vsGP5wJRWRI1rxk26FQwftau7sIw7SA6NRQoEGwltGRaYvA+Cdp/wO3KNAG+EkMRD0GoRYXMHTT7EqYKTs6uOEv6Vs4ixnzZe1bHo0HTtGq80tInad6LTYu/HkGpgvFyiO3Aq2aJ9H3eh9Od59Enm8RhRpNvdJk0qAK7Ei+dUZBM0R8geLdelpC33OvYYFmBrYagDbKOTpA933IDxqjON9Ps0W4B3P9ZpOP+zBO1yz02wmsCXmzZv5iH9l6f2cqACkkV4XFT+mC4u0iyg93RjknXP4beoMrndmv6xFSYpZxY5nOOQ+NCqGjfMomA/tsEcUiecnY2BrgZ/LTGRQ3zK0kXZkofzePnm7hRLOcxppZHCEHAwXB9LT4djfGrhMmrB9qV/zmKn4zyLeBCUT3sZ00uziYR0cv6FaAkYpj+HV0EkVTJupWGfNv9wqrZhrRjv11S/OlkcghMvMN22mYV1CaJct9s009hWvK/hHtqpayrkwYAR7iUwg4Ge/JvB3XrN4RA1Leunr7ETESSn5q7t/ntHeKaIo+dAr22D0HI1AQrs2zAKGeynYknI9pWr8mH8RPVZvmQSaLTzuYx2kYTwokbtr8iFF3Zz0bnxS/ok9xCtpw6RhbcQOblB7H4OV2zMH29dt5diM7fkDxJh7a4MVB/T0q2FeA+Hg9Twp8CuO0+nHEIXhHQnDPVV66IQT29D6iXj5tMru+7W+lIHccn6VB+0bv0kX/6DJLcYWvpy2fPwnVSsrABCjKjdFT3NRpsXj5PKr3MhjlK4ZVm70oSMxmUXfsx0Pkjco4feXF21hjs6lPqv1QtJx68fe82xifJlDO2TGL4le3UWG0X4yBoqyO6sUf1+g7iJ8wzudXMXdJ2AIWzZWSVqO53Cqq99MmvSJyRXvTEXppEFqlgAmfOAVFrJG9X8egsEpw30+bgzv7g5c7XNtcBF4h0mU8Xtjpe5GP03P0d0zB3OMqMlN8nbP7Z2907w2laga9zwr5fBtqnamEMvYaUN5FepR06MsSyj85nTouSjt+HmFms2ONgyokRarRbc26qaxtYdfNS59+VApcVVXym7M27BZpmITDgfGrz71bdOs08BamT7t9ZdE2Rves0ATh/ncOy9vm5ofdQK3RuCi1kZMhHPbRxACmdEzMWMJe2waQbH02u1VLN3v1RPJrp4N1S+RDgtDduz6y2Msgn55R7IUrWVci8DMisL/t08/IuOxaItAqBKRRtApZSffQIiCN4tCKTjLeKgSkUbQKWUn30CIgjeLQik4y3ioEpFG0CllJ99AiII3i0IpOMt4qBKRRtApZSffQIiCN4tCKTjLeKgSkUbQKWUn30CIgjeLQik4y3ioE/gcm0wHHxPR93gAAAABJRU5ErkJggg==`,
		},
	},
	Images: []EmailAttachment{
		{
			Name:          `Test Image`,
			MimeType:      `image/png`,
			Base64Content: `data:image/png;base64,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`,
		},
	},
	Tags: []string{
		`test 1`,
		`test 2`,
	},
	Metadata: map[string]string{
		`meta1`: `meta1_val`,
		`meta2`: `meta2_val`,
	},
}

var testParams = SendParams{
	// should do a separate test for async
	SendAsync:   false,
	TrackOpens:  true,
	TrackClicks: true,
}

func initData() (*mandrill, error) {

	sender := &MailRecipient{
		Name:  `Mandrill From Name`,
		Email: testFromEmail,
	}

	m, err := NewMandrill(mandrillTestKey, "globio.co", sender, new(http.Client))
	if err != nil {
		errors.New("Failed to create new Mandrill instance")
	}

	return m, nil
}

func TestBulk_Mail(t *testing.T) {

	m, err := initData()
	if err != nil {
		t.Error(err.Error())
	}

	tmpl, err := template.New(`mail_test`).Parse(testTemplate)
	if err != nil {
		t.Error(err.Error())
	}
	testMessage.HTMLTemplate = tmpl

	response, err := m.BulkMail(testRecipients, &testMessage, &testParams)
	if err != nil {
		t.Errorf("Mail failed with error : %s", err.Error())
	}

	if len(response) != len(testRecipients) {
		t.Errorf("BulkMail send mail for %d recipients, but only received %d responses", len(response), len(testRecipients))
	}

	for _, v := range response {
		if !(v.Status == MAIL_MESSAGE_SENT || v.Status == MAIL_MESSAGE_QUEUED) {
			t.Errorf(`BulkMail to %s failed with error: %s`, v.Email, v.Error)
		}
	}
}

func TestMandrill_TemplateMail(t *testing.T) {

	m, err := initData()
	if err != nil {
		t.Error(err.Error())
	}

	tmpl, err := template.New(`mail_test`).Parse(testTemplate)
	if err != nil {
		t.Error(err.Error())
	}
	testMessage.HTMLTemplate = tmpl

	response, err := m.TemplateMail(testToFirstEmail, `Test Mail`, tmpl, map[string]string{
		`TestParam`: `This is a test!`,
	})

	if err != nil {
		t.Errorf("TemplateMail failed with error : %s", err.Error())
	}

	if !(response.Status == MAIL_MESSAGE_SENT || response.Status == MAIL_MESSAGE_QUEUED) {
		t.Errorf(`TemplateMail to %s failed with error: %s`, response.Email, response.Error)
	}
}

func TestMandrill_SimpleMail(t *testing.T) {

	m, err := initData()
	if err != nil {
		t.Error(err.Error())
	}

	response, err := m.SimpleMail(testToFirstEmail, `Test Mail`, `simple_mail_test`, `This is a test!`)

	if err != nil {
		t.Errorf("SimpleMail failed with error : %s", err.Error())
	}

	if !(response.Status == MAIL_MESSAGE_SENT || response.Status == MAIL_MESSAGE_QUEUED) {
		t.Errorf(`SimpleMail to %s failed with error: %s`, response.Email, response.Error)
	}
}
