<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Exchange    </title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta1/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-giJF6kkoqNQ00vy+HMDP7azOuL0xtbfIcaT9wjKHr8RbDVddVHyTfAAsrekwKmP1" crossorigin="anonymous">
    <script src="js/script.js"></script>

</head>
<body>
<form id="contact-form" method="post" action="inc/process.php" role="form">

    <div class="messages"></div>

    <div class="controls">

        <div class="row">
            <div class="col-md-6">
                <div class="form-group">
                    <label for="form_name">Firstname *</label>
                    <input id="form_name" type="text" name="name" class="form-control" placeholder="Please enter your name*" required="required" data-error="Firstname is required.">
                    <div class="help-block with-errors"></div>
                </div>
            </div>
            <div class="col-md-6">
                <div class="form-group">
                    <label for="form_lastname">Email *</label>
                    <input id="form_email" type="email" name="email" class="form-control" placeholder="Please enter your email *" required="required" data-error="email is required.">
                    <div class="help-block with-errors"></div>
                </div>
            </div>
        </div>
        <div class="row">
            <div class="col-md-4">
                <div class="form-group">
                    <label for="form_other_firstname">Their Firstname</label>
                    <input id="form_other_firstname" type="text" name="other_firstname" class="form-control" placeholder="Please enter their firstname*" required="required" data-error="Valid firstname is required.">
                    <div class="help-block with-errors"></div>
                </div>
            </div>
            <div class="col-md-4">
                <div class="form-group">
                    <label for="form_other_lastname">Their lastnamae</label>
                    <input id="form_other_lastname" type="text" name="other_lastname" class="form-control" placeholder="Please enter their lastname*" required="required" data-error="Valid last is required.">
                    <div class="help-block with-errors"></div>
                </div>
            </div>
            <div class="col-md-4">
                <div class="form-group">
                    <label for="form_other_email">Their email</label>
                    <input id="form_other_email" type="email" name="other_email" class="form-control" placeholder="Please enter their Email*" required="required" data-error="Valid Email is required.">
                    <div class="help-block with-errors"></div>
                </div>
            </div>
            
            
        </div>
        <div class="row">
            <div class="col-md-12">
                <div class="form-group">
                    <label for="form_message">Message *</label>
                    <textarea id="form_message" name="message" class="form-control" placeholder="Message for me *" rows="4" required="required" data-error="Please, leave us a message."></textarea>
                    <div class="help-block with-errors"></div>
                </div>
            </div>
            </div>

            <div id="success"></div>
                        <div class="row">
                            <div  class="form-group col-xs-12">
                            <div class="g-recaptcha" data-sitekey="6LeGDHwaAAAAADAGSax5zzuM16NPJmjKOX5jOgP9"></div>
                                <button id="contactFormSubmit"   type="submit" class="btn btn-success btn-lg">Send</button>
                            <div id="wasitasuccess"> </div>
                            </div>

                        </div>
        <div class="row">
            <div class="col-md-12">
                <p class="text-muted">
                    <strong>*</strong> These fields are required. Contact form template by
                    <a href="https://bootstrapious.com/p/how-to-build-a-working-bootstrap-contact-form" target="_blank">Bootstrapious</a>.</p>
            </div>
        </div>
    </div>

</form>
<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta1/dist/js/bootstrap.bundle.min.js" integrity="sha384-ygbV9kiqUc6oa4msXn9868pTtWMgiQaeYH7/t7LECLbyPA2x65Kgf80OJFdroafW" crossorigin="anonymous"></script>
<script src="https://www.google.com/recaptcha/api.js"></script>

</body>
</html>