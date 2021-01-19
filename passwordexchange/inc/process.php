<?php
//namespace SendGrid;
require (__DIR__ . '/../vendor/autoload.php');

$dotenv = Dotenv\Dotenv::createImmutable(__DIR__ . '/../');
$dotenv->load();

use \Aws\Ses\SesClient;
use \Aws\Credentials\CredentialProvider;
use \Aws\Exception\AwsException;


function sendEmail(){
	### TODO ### 
    # Create function with all variables in an array #
	$provider = CredentialProvider::env();
echo '<?xml version="1.0" encoding="UTF-8" ?>'; 
$SesClient = new SesClient([
    'version' => 'latest',
	'region'  => 'us-west-2',
	'credentials' => $provider
]);
$sender_email = 'anthony@anthony.bible';

	
		$recieverEmail=[$_POST["email"], "anthony@anthonybible.com"];
		$receiverid = $_POST["name"];
		$phone=$_POST["phone"];
		$message = $_POST["message"];
		$subject = "Thanks for contacting me";
		$plaintext_body = 'This email was sent with Amazon SES using the AWS SDK for PHP.' ;
		$html_body =  '<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
		<html xmlns="http://www.w3.org/1999/xhtml">
		
		<head style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
		  <meta http-equiv="Content-Type" content="text/html; charset=utf-8" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
		  <meta name="viewport" content="width=device-width, initial-scale=1, minimum-scale=1, maximum-scale=1" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
		  <!--[if !mso]><!-->
		  <meta http-equiv="X-UA-Compatible" content="IE=Edge" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
		  <!--<![endif]-->
		  <!--[if (gte mso 9)|(IE)]>
			<xml>
			<o:OfficeDocumentSettings>
			<o:AllowPNG/>
			<o:PixelsPerInch>96</o:PixelsPerInch>
			</o:OfficeDocumentSettings>
			</xml>
			<![endif]-->
		  <!--[if (gte mso 9)|(IE)]>
			<style type="text/css">
			  body {width: 600px;margin: 0 auto;}
			  table {border-collapse: collapse;}
			  table, td {mso-table-lspace: 0pt;mso-table-rspace: 0pt;}
			  img {-ms-interpolation-mode: bicubic;}
			</style>
			<![endif]-->
		
		
		  <!--user entered Head Start-->
		
		
		
		
		
		
		  <!--End Head user entered-->
		
		</head>
		
		<body style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: #fff; box-sizing: border-box; color: #333; font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; font-size: 14px; line-height: 1.42857143; margin: 0;">
		  <style type="text/css" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
			@media screen and (max-width:480px) {
			  .preheader .rightColumnContent,
			  .footer .rightColumnContent {
				text-align: left !important;
			  }
			  .preheader .rightColumnContent div,
			  .preheader .rightColumnContent span,
			  .footer .rightColumnContent div,
			  .footer .rightColumnContent span {
				text-align: left !important;
			  }
			  .preheader .rightColumnContent,
			  .preheader .leftColumnContent {
				font-size: 80% !important;
				padding: 5px 0;
			  }
			  table.wrapper-mobile {
				width: 100% !important;
				table-layout: fixed;
			  }
			  img.max-width {
				height: auto !important;
				max-width: 480px !important;
			  }
			  a.bulletproof-button {
				display: block !important;
				width: auto !important;
				font-size: 80%;
				padding-left: 0 !important;
				padding-right: 0 !important;
			  }
			  .columns {
				width: 100% !important;
			  }
			  .column {
				display: block !important;
				width: 100% !important;
				padding-left: 0 !important;
				padding-right: 0 !important;
				margin-left: 0 !important;
				margin-right: 0 !important;
			  }
			  .total_spacer {
				padding: 0px 0px 0px 0px;
			  }
			}
			
			@media print {
			  *,
			  :after,
			  :before {
				color: #000!important;
				text-shadow: none!important;
				background: 0 0!important;
				-webkit-box-shadow: none!important;
				box-shadow: none!important;
			  }
			  a,
			  a:visited {
				text-decoration: underline;
			  }
			  a[href]:after {
				content: " (" attr(href) ")";
			  }
			  abbr[title]:after {
				content: " (" attr(title) ")";
			  }
			  a[href^="javascript:"]:after,
			  a[href^="#"]:after {
				content: "";
			  }
			  blockquote,
			  pre {
				border: 1px solid #999;
				page-break-inside: avoid;
			  }
			  thead {
				display: table-header-group;
			  }
			  img,
			  tr {
				page-break-inside: avoid;
			  }
			  img {
				max-width: 100%!important;
			  }
			  h2,
			  h3,
			  p {
				orphans: 3;
				widows: 3;
			  }
			  h2,
			  h3 {
				page-break-after: avoid;
			  }
			  .navbar {
				display: none;
			  }
			  .btn>.caret,
			  .dropup>.btn>.caret {
				border-top-color: #000!important;
			  }
			  .label {
				border: 1px solid #000;
			  }
			  .table {
				border-collapse: collapse!important;
			  }
			  .table td,
			  .table th {
				background-color: #fff!important;
			  }
			  .table-bordered td,
			  .table-bordered th {
				border: 1px solid #ddd!important;
			  }
			}
			
			@media (min-width:768px) {
			  .lead {
				font-size: 21px;
			  }
			}
			
			@media (min-width:768px) {
			  .dl-horizontal dt {
				float: left;
				width: 160px;
				overflow: hidden;
				clear: left;
				text-align: right;
				text-overflow: ellipsis;
				white-space: nowrap;
			  }
			  .dl-horizontal dd {
				margin-left: 180px;
			  }
			}
			
			@media (min-width:768px) {
			  .container {
				width: 750px;
			  }
			}
			
			@media (min-width:992px) {
			  .container {
				width: 970px;
			  }
			}
			
			@media (min-width:1200px) {
			  .container {
				width: 1170px;
			  }
			}
			
			@media (min-width:768px) {
			  .col-sm-1,
			  .col-sm-10,
			  .col-sm-11,
			  .col-sm-12,
			  .col-sm-2,
			  .col-sm-3,
			  .col-sm-4,
			  .col-sm-5,
			  .col-sm-6,
			  .col-sm-7,
			  .col-sm-8,
			  .col-sm-9 {
				float: left;
			  }
			  .col-sm-12 {
				width: 100%;
			  }
			  .col-sm-11 {
				width: 91.66666667%;
			  }
			  .col-sm-10 {
				width: 83.33333333%;
			  }
			  .col-sm-9 {
				width: 75%;
			  }
			  .col-sm-8 {
				width: 66.66666667%;
			  }
			  .col-sm-7 {
				width: 58.33333333%;
			  }
			  .col-sm-6 {
				width: 50%;
			  }
			  .col-sm-5 {
				width: 41.66666667%;
			  }
			  .col-sm-4 {
				width: 33.33333333%;
			  }
			  .col-sm-3 {
				width: 25%;
			  }
			  .col-sm-2 {
				width: 16.66666667%;
			  }
			  .col-sm-1 {
				width: 8.33333333%;
			  }
			  .col-sm-pull-12 {
				right: 100%;
			  }
			  .col-sm-pull-11 {
				right: 91.66666667%;
			  }
			  .col-sm-pull-10 {
				right: 83.33333333%;
			  }
			  .col-sm-pull-9 {
				right: 75%;
			  }
			  .col-sm-pull-8 {
				right: 66.66666667%;
			  }
			  .col-sm-pull-7 {
				right: 58.33333333%;
			  }
			  .col-sm-pull-6 {
				right: 50%;
			  }
			  .col-sm-pull-5 {
				right: 41.66666667%;
			  }
			  .col-sm-pull-4 {
				right: 33.33333333%;
			  }
			  .col-sm-pull-3 {
				right: 25%;
			  }
			  .col-sm-pull-2 {
				right: 16.66666667%;
			  }
			  .col-sm-pull-1 {
				right: 8.33333333%;
			  }
			  .col-sm-pull-0 {
				right: auto;
			  }
			  .col-sm-push-12 {
				left: 100%;
			  }
			  .col-sm-push-11 {
				left: 91.66666667%;
			  }
			  .col-sm-push-10 {
				left: 83.33333333%;
			  }
			  .col-sm-push-9 {
				left: 75%;
			  }
			  .col-sm-push-8 {
				left: 66.66666667%;
			  }
			  .col-sm-push-7 {
				left: 58.33333333%;
			  }
			  .col-sm-push-6 {
				left: 50%;
			  }
			  .col-sm-push-5 {
				left: 41.66666667%;
			  }
			  .col-sm-push-4 {
				left: 33.33333333%;
			  }
			  .col-sm-push-3 {
				left: 25%;
			  }
			  .col-sm-push-2 {
				left: 16.66666667%;
			  }
			  .col-sm-push-1 {
				left: 8.33333333%;
			  }
			  .col-sm-push-0 {
				left: auto;
			  }
			  .col-sm-offset-12 {
				margin-left: 100%;
			  }
			  .col-sm-offset-11 {
				margin-left: 91.66666667%;
			  }
			  .col-sm-offset-10 {
				margin-left: 83.33333333%;
			  }
			  .col-sm-offset-9 {
				margin-left: 75%;
			  }
			  .col-sm-offset-8 {
				margin-left: 66.66666667%;
			  }
			  .col-sm-offset-7 {
				margin-left: 58.33333333%;
			  }
			  .col-sm-offset-6 {
				margin-left: 50%;
			  }
			  .col-sm-offset-5 {
				margin-left: 41.66666667%;
			  }
			  .col-sm-offset-4 {
				margin-left: 33.33333333%;
			  }
			  .col-sm-offset-3 {
				margin-left: 25%;
			  }
			  .col-sm-offset-2 {
				margin-left: 16.66666667%;
			  }
			  .col-sm-offset-1 {
				margin-left: 8.33333333%;
			  }
			  .col-sm-offset-0 {
				margin-left: 0;
			  }
			}
			
			@media (min-width:992px) {
			  .col-md-1,
			  .col-md-10,
			  .col-md-11,
			  .col-md-12,
			  .col-md-2,
			  .col-md-3,
			  .col-md-4,
			  .col-md-5,
			  .col-md-6,
			  .col-md-7,
			  .col-md-8,
			  .col-md-9 {
				float: left;
			  }
			  .col-md-12 {
				width: 100%;
			  }
			  .col-md-11 {
				width: 91.66666667%;
			  }
			  .col-md-10 {
				width: 83.33333333%;
			  }
			  .col-md-9 {
				width: 75%;
			  }
			  .col-md-8 {
				width: 66.66666667%;
			  }
			  .col-md-7 {
				width: 58.33333333%;
			  }
			  .col-md-6 {
				width: 50%;
			  }
			  .col-md-5 {
				width: 41.66666667%;
			  }
			  .col-md-4 {
				width: 33.33333333%;
			  }
			  .col-md-3 {
				width: 25%;
			  }
			  .col-md-2 {
				width: 16.66666667%;
			  }
			  .col-md-1 {
				width: 8.33333333%;
			  }
			  .col-md-pull-12 {
				right: 100%;
			  }
			  .col-md-pull-11 {
				right: 91.66666667%;
			  }
			  .col-md-pull-10 {
				right: 83.33333333%;
			  }
			  .col-md-pull-9 {
				right: 75%;
			  }
			  .col-md-pull-8 {
				right: 66.66666667%;
			  }
			  .col-md-pull-7 {
				right: 58.33333333%;
			  }
			  .col-md-pull-6 {
				right: 50%;
			  }
			  .col-md-pull-5 {
				right: 41.66666667%;
			  }
			  .col-md-pull-4 {
				right: 33.33333333%;
			  }
			  .col-md-pull-3 {
				right: 25%;
			  }
			  .col-md-pull-2 {
				right: 16.66666667%;
			  }
			  .col-md-pull-1 {
				right: 8.33333333%;
			  }
			  .col-md-pull-0 {
				right: auto;
			  }
			  .col-md-push-12 {
				left: 100%;
			  }
			  .col-md-push-11 {
				left: 91.66666667%;
			  }
			  .col-md-push-10 {
				left: 83.33333333%;
			  }
			  .col-md-push-9 {
				left: 75%;
			  }
			  .col-md-push-8 {
				left: 66.66666667%;
			  }
			  .col-md-push-7 {
				left: 58.33333333%;
			  }
			  .col-md-push-6 {
				left: 50%;
			  }
			  .col-md-push-5 {
				left: 41.66666667%;
			  }
			  .col-md-push-4 {
				left: 33.33333333%;
			  }
			  .col-md-push-3 {
				left: 25%;
			  }
			  .col-md-push-2 {
				left: 16.66666667%;
			  }
			  .col-md-push-1 {
				left: 8.33333333%;
			  }
			  .col-md-push-0 {
				left: auto;
			  }
			  .col-md-offset-12 {
				margin-left: 100%;
			  }
			  .col-md-offset-11 {
				margin-left: 91.66666667%;
			  }
			  .col-md-offset-10 {
				margin-left: 83.33333333%;
			  }
			  .col-md-offset-9 {
				margin-left: 75%;
			  }
			  .col-md-offset-8 {
				margin-left: 66.66666667%;
			  }
			  .col-md-offset-7 {
				margin-left: 58.33333333%;
			  }
			  .col-md-offset-6 {
				margin-left: 50%;
			  }
			  .col-md-offset-5 {
				margin-left: 41.66666667%;
			  }
			  .col-md-offset-4 {
				margin-left: 33.33333333%;
			  }
			  .col-md-offset-3 {
				margin-left: 25%;
			  }
			  .col-md-offset-2 {
				margin-left: 16.66666667%;
			  }
			  .col-md-offset-1 {
				margin-left: 8.33333333%;
			  }
			  .col-md-offset-0 {
				margin-left: 0;
			  }
			}
			
			@media (min-width:1200px) {
			  .col-lg-1,
			  .col-lg-10,
			  .col-lg-11,
			  .col-lg-12,
			  .col-lg-2,
			  .col-lg-3,
			  .col-lg-4,
			  .col-lg-5,
			  .col-lg-6,
			  .col-lg-7,
			  .col-lg-8,
			  .col-lg-9 {
				float: left;
			  }
			  .col-lg-12 {
				width: 100%;
			  }
			  .col-lg-11 {
				width: 91.66666667%;
			  }
			  .col-lg-10 {
				width: 83.33333333%;
			  }
			  .col-lg-9 {
				width: 75%;
			  }
			  .col-lg-8 {
				width: 66.66666667%;
			  }
			  .col-lg-7 {
				width: 58.33333333%;
			  }
			  .col-lg-6 {
				width: 50%;
			  }
			  .col-lg-5 {
				width: 41.66666667%;
			  }
			  .col-lg-4 {
				width: 33.33333333%;
			  }
			  .col-lg-3 {
				width: 25%;
			  }
			  .col-lg-2 {
				width: 16.66666667%;
			  }
			  .col-lg-1 {
				width: 8.33333333%;
			  }
			  .col-lg-pull-12 {
				right: 100%;
			  }
			  .col-lg-pull-11 {
				right: 91.66666667%;
			  }
			  .col-lg-pull-10 {
				right: 83.33333333%;
			  }
			  .col-lg-pull-9 {
				right: 75%;
			  }
			  .col-lg-pull-8 {
				right: 66.66666667%;
			  }
			  .col-lg-pull-7 {
				right: 58.33333333%;
			  }
			  .col-lg-pull-6 {
				right: 50%;
			  }
			  .col-lg-pull-5 {
				right: 41.66666667%;
			  }
			  .col-lg-pull-4 {
				right: 33.33333333%;
			  }
			  .col-lg-pull-3 {
				right: 25%;
			  }
			  .col-lg-pull-2 {
				right: 16.66666667%;
			  }
			  .col-lg-pull-1 {
				right: 8.33333333%;
			  }
			  .col-lg-pull-0 {
				right: auto;
			  }
			  .col-lg-push-12 {
				left: 100%;
			  }
			  .col-lg-push-11 {
				left: 91.66666667%;
			  }
			  .col-lg-push-10 {
				left: 83.33333333%;
			  }
			  .col-lg-push-9 {
				left: 75%;
			  }
			  .col-lg-push-8 {
				left: 66.66666667%;
			  }
			  .col-lg-push-7 {
				left: 58.33333333%;
			  }
			  .col-lg-push-6 {
				left: 50%;
			  }
			  .col-lg-push-5 {
				left: 41.66666667%;
			  }
			  .col-lg-push-4 {
				left: 33.33333333%;
			  }
			  .col-lg-push-3 {
				left: 25%;
			  }
			  .col-lg-push-2 {
				left: 16.66666667%;
			  }
			  .col-lg-push-1 {
				left: 8.33333333%;
			  }
			  .col-lg-push-0 {
				left: auto;
			  }
			  .col-lg-offset-12 {
				margin-left: 100%;
			  }
			  .col-lg-offset-11 {
				margin-left: 91.66666667%;
			  }
			  .col-lg-offset-10 {
				margin-left: 83.33333333%;
			  }
			  .col-lg-offset-9 {
				margin-left: 75%;
			  }
			  .col-lg-offset-8 {
				margin-left: 66.66666667%;
			  }
			  .col-lg-offset-7 {
				margin-left: 58.33333333%;
			  }
			  .col-lg-offset-6 {
				margin-left: 50%;
			  }
			  .col-lg-offset-5 {
				margin-left: 41.66666667%;
			  }
			  .col-lg-offset-4 {
				margin-left: 33.33333333%;
			  }
			  .col-lg-offset-3 {
				margin-left: 25%;
			  }
			  .col-lg-offset-2 {
				margin-left: 16.66666667%;
			  }
			  .col-lg-offset-1 {
				margin-left: 8.33333333%;
			  }
			  .col-lg-offset-0 {
				margin-left: 0;
			  }
			}
			
			@media screen and (max-width:767px) {
			  .table-responsive {
				width: 100%;
				margin-bottom: 15px;
				overflow-y: hidden;
				-ms-overflow-style: -ms-autohiding-scrollbar;
				border: 1px solid #ddd;
			  }
			  .table-responsive>.table {
				margin-bottom: 0;
			  }
			  .table-responsive>.table>tbody>tr>td,
			  .table-responsive>.table>tbody>tr>th,
			  .table-responsive>.table>tfoot>tr>td,
			  .table-responsive>.table>tfoot>tr>th,
			  .table-responsive>.table>thead>tr>td,
			  .table-responsive>.table>thead>tr>th {
				white-space: nowrap;
			  }
			  .table-responsive>.table-bordered {
				border: 0;
			  }
			  .table-responsive>.table-bordered>tbody>tr>td:first-child,
			  .table-responsive>.table-bordered>tbody>tr>th:first-child,
			  .table-responsive>.table-bordered>tfoot>tr>td:first-child,
			  .table-responsive>.table-bordered>tfoot>tr>th:first-child,
			  .table-responsive>.table-bordered>thead>tr>td:first-child,
			  .table-responsive>.table-bordered>thead>tr>th:first-child {
				border-left: 0;
			  }
			  .table-responsive>.table-bordered>tbody>tr>td:last-child,
			  .table-responsive>.table-bordered>tbody>tr>th:last-child,
			  .table-responsive>.table-bordered>tfoot>tr>td:last-child,
			  .table-responsive>.table-bordered>tfoot>tr>th:last-child,
			  .table-responsive>.table-bordered>thead>tr>td:last-child,
			  .table-responsive>.table-bordered>thead>tr>th:last-child {
				border-right: 0;
			  }
			  .table-responsive>.table-bordered>tbody>tr:last-child>td,
			  .table-responsive>.table-bordered>tbody>tr:last-child>th,
			  .table-responsive>.table-bordered>tfoot>tr:last-child>td,
			  .table-responsive>.table-bordered>tfoot>tr:last-child>th {
				border-bottom: 0;
			  }
			}
			
			@media screen and (-webkit-min-device-pixel-ratio:0) {
			  input[type=date].form-control,
			  input[type=time].form-control,
			  input[type=datetime-local].form-control,
			  input[type=month].form-control {
				line-height: 34px;
			  }
			  .input-group-sm input[type=date],
			  .input-group-sm input[type=time],
			  .input-group-sm input[type=datetime-local],
			  .input-group-sm input[type=month],
			  input[type=date].input-sm,
			  input[type=time].input-sm,
			  input[type=datetime-local].input-sm,
			  input[type=month].input-sm {
				line-height: 30px;
			  }
			  .input-group-lg input[type=date],
			  .input-group-lg input[type=time],
			  .input-group-lg input[type=datetime-local],
			  .input-group-lg input[type=month],
			  input[type=date].input-lg,
			  input[type=time].input-lg,
			  input[type=datetime-local].input-lg,
			  input[type=month].input-lg {
				line-height: 46px;
			  }
			}
			
			@media (min-width:768px) {
			  .form-inline .form-group {
				display: inline-block;
				margin-bottom: 0;
				vertical-align: middle;
			  }
			  .form-inline .form-control {
				display: inline-block;
				width: auto;
				vertical-align: middle;
			  }
			  .form-inline .form-control-static {
				display: inline-block;
			  }
			  .form-inline .input-group {
				display: inline-table;
				vertical-align: middle;
			  }
			  .form-inline .input-group .form-control,
			  .form-inline .input-group .input-group-addon,
			  .form-inline .input-group .input-group-btn {
				width: auto;
			  }
			  .form-inline .input-group>.form-control {
				width: 100%;
			  }
			  .form-inline .control-label {
				margin-bottom: 0;
				vertical-align: middle;
			  }
			  .form-inline .checkbox,
			  .form-inline .radio {
				display: inline-block;
				margin-top: 0;
				margin-bottom: 0;
				vertical-align: middle;
			  }
			  .form-inline .checkbox label,
			  .form-inline .radio label {
				padding-left: 0;
			  }
			  .form-inline .checkbox input[type=checkbox],
			  .form-inline .radio input[type=radio] {
				position: relative;
				margin-left: 0;
			  }
			  .form-inline .has-feedback .form-control-feedback {
				top: 0;
			  }
			}
			
			@media (min-width:768px) {
			  .form-horizontal .control-label {
				padding-top: 7px;
				margin-bottom: 0;
				text-align: right;
			  }
			}
			
			@media (min-width:768px) {
			  .form-horizontal .form-group-lg .control-label {
				padding-top: 11px;
				font-size: 18px;
			  }
			}
			
			@media (min-width:768px) {
			  .form-horizontal .form-group-sm .control-label {
				padding-top: 6px;
				font-size: 12px;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar-right .dropdown-menu {
				right: 0;
				left: auto;
			  }
			  .navbar-right .dropdown-menu-left {
				right: auto;
				left: 0;
			  }
			}
			
			@media (min-width:768px) {
			  .nav-tabs.nav-justified>li {
				display: table-cell;
				width: 1%;
			  }
			  .nav-tabs.nav-justified>li>a {
				margin-bottom: 0;
			  }
			}
			
			@media (min-width:768px) {
			  .nav-tabs.nav-justified>li>a {
				border-bottom: 1px solid #ddd;
				border-radius: 4px 4px 0 0;
			  }
			  .nav-tabs.nav-justified>.active>a,
			  .nav-tabs.nav-justified>.active>a:focus,
			  .nav-tabs.nav-justified>.active>a:hover {
				border-bottom-color: #fff;
			  }
			}
			
			@media (min-width:768px) {
			  .nav-justified>li {
				display: table-cell;
				width: 1%;
			  }
			  .nav-justified>li>a {
				margin-bottom: 0;
			  }
			}
			
			@media (min-width:768px) {
			  .nav-tabs-justified>li>a {
				border-bottom: 1px solid #ddd;
				border-radius: 4px 4px 0 0;
			  }
			  .nav-tabs-justified>.active>a,
			  .nav-tabs-justified>.active>a:focus,
			  .nav-tabs-justified>.active>a:hover {
				border-bottom-color: #fff;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar {
				border-radius: 4px;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar-header {
				float: left;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar-collapse {
				width: auto;
				border-top: 0;
				-webkit-box-shadow: none;
				box-shadow: none;
			  }
			  .navbar-collapse.collapse {
				display: block!important;
				height: auto!important;
				padding-bottom: 0;
				overflow: visible!important;
			  }
			  .navbar-collapse.in {
				overflow-y: visible;
			  }
			  .navbar-fixed-bottom .navbar-collapse,
			  .navbar-fixed-top .navbar-collapse,
			  .navbar-static-top .navbar-collapse {
				padding-right: 0;
				padding-left: 0;
			  }
			}
			
			@media (max-device-width:480px) and (orientation:landscape) {
			  .navbar-fixed-bottom .navbar-collapse,
			  .navbar-fixed-top .navbar-collapse {
				max-height: 200px;
			  }
			}
			
			@media (min-width:768px) {
			  .container-fluid>.navbar-collapse,
			  .container-fluid>.navbar-header,
			  .container>.navbar-collapse,
			  .container>.navbar-header {
				margin-right: 0;
				margin-left: 0;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar-static-top {
				border-radius: 0;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar-fixed-bottom,
			  .navbar-fixed-top {
				border-radius: 0;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar>.container .navbar-brand,
			  .navbar>.container-fluid .navbar-brand {
				margin-left: -15px;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar-toggle {
				display: none;
			  }
			}
			
			@media (max-width:767px) {
			  .navbar-nav .open .dropdown-menu {
				position: static;
				float: none;
				width: auto;
				margin-top: 0;
				background-color: transparent;
				border: 0;
				-webkit-box-shadow: none;
				box-shadow: none;
			  }
			  .navbar-nav .open .dropdown-menu .dropdown-header,
			  .navbar-nav .open .dropdown-menu>li>a {
				padding: 5px 15px 5px 25px;
			  }
			  .navbar-nav .open .dropdown-menu>li>a {
				line-height: 20px;
			  }
			  .navbar-nav .open .dropdown-menu>li>a:focus,
			  .navbar-nav .open .dropdown-menu>li>a:hover {
				background-image: none;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar-nav {
				float: left;
				margin: 0;
			  }
			  .navbar-nav>li {
				float: left;
			  }
			  .navbar-nav>li>a {
				padding-top: 15px;
				padding-bottom: 15px;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar-form .form-group {
				display: inline-block;
				margin-bottom: 0;
				vertical-align: middle;
			  }
			  .navbar-form .form-control {
				display: inline-block;
				width: auto;
				vertical-align: middle;
			  }
			  .navbar-form .form-control-static {
				display: inline-block;
			  }
			  .navbar-form .input-group {
				display: inline-table;
				vertical-align: middle;
			  }
			  .navbar-form .input-group .form-control,
			  .navbar-form .input-group .input-group-addon,
			  .navbar-form .input-group .input-group-btn {
				width: auto;
			  }
			  .navbar-form .input-group>.form-control {
				width: 100%;
			  }
			  .navbar-form .control-label {
				margin-bottom: 0;
				vertical-align: middle;
			  }
			  .navbar-form .checkbox,
			  .navbar-form .radio {
				display: inline-block;
				margin-top: 0;
				margin-bottom: 0;
				vertical-align: middle;
			  }
			  .navbar-form .checkbox label,
			  .navbar-form .radio label {
				padding-left: 0;
			  }
			  .navbar-form .checkbox input[type=checkbox],
			  .navbar-form .radio input[type=radio] {
				position: relative;
				margin-left: 0;
			  }
			  .navbar-form .has-feedback .form-control-feedback {
				top: 0;
			  }
			}
			
			@media (max-width:767px) {
			  .navbar-form .form-group {
				margin-bottom: 5px;
			  }
			  .navbar-form .form-group:last-child {
				margin-bottom: 0;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar-form {
				width: auto;
				padding-top: 0;
				padding-bottom: 0;
				margin-right: 0;
				margin-left: 0;
				border: 0;
				-webkit-box-shadow: none;
				box-shadow: none;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar-text {
				float: left;
				margin-right: 15px;
				margin-left: 15px;
			  }
			}
			
			@media (min-width:768px) {
			  .navbar-left {
				float: left!important;
			  }
			  .navbar-right {
				float: right!important;
				margin-right: -15px;
			  }
			  .navbar-right~.navbar-right {
				margin-right: 0;
			  }
			}
			
			@media (max-width:767px) {
			  .navbar-default .navbar-nav .open .dropdown-menu>li>a {
				color: #777;
			  }
			  .navbar-default .navbar-nav .open .dropdown-menu>li>a:focus,
			  .navbar-default .navbar-nav .open .dropdown-menu>li>a:hover {
				color: #333;
				background-color: transparent;
			  }
			  .navbar-default .navbar-nav .open .dropdown-menu>.active>a,
			  .navbar-default .navbar-nav .open .dropdown-menu>.active>a:focus,
			  .navbar-default .navbar-nav .open .dropdown-menu>.active>a:hover {
				color: #555;
				background-color: #e7e7e7;
			  }
			  .navbar-default .navbar-nav .open .dropdown-menu>.disabled>a,
			  .navbar-default .navbar-nav .open .dropdown-menu>.disabled>a:focus,
			  .navbar-default .navbar-nav .open .dropdown-menu>.disabled>a:hover {
				color: #ccc;
				background-color: transparent;
			  }
			}
			
			@media (max-width:767px) {
			  .navbar-inverse .navbar-nav .open .dropdown-menu>.dropdown-header {
				border-color: #080808;
			  }
			  .navbar-inverse .navbar-nav .open .dropdown-menu .divider {
				background-color: #080808;
			  }
			  .navbar-inverse .navbar-nav .open .dropdown-menu>li>a {
				color: #9d9d9d;
			  }
			  .navbar-inverse .navbar-nav .open .dropdown-menu>li>a:focus,
			  .navbar-inverse .navbar-nav .open .dropdown-menu>li>a:hover {
				color: #fff;
				background-color: transparent;
			  }
			  .navbar-inverse .navbar-nav .open .dropdown-menu>.active>a,
			  .navbar-inverse .navbar-nav .open .dropdown-menu>.active>a:focus,
			  .navbar-inverse .navbar-nav .open .dropdown-menu>.active>a:hover {
				color: #fff;
				background-color: #080808;
			  }
			  .navbar-inverse .navbar-nav .open .dropdown-menu>.disabled>a,
			  .navbar-inverse .navbar-nav .open .dropdown-menu>.disabled>a:focus,
			  .navbar-inverse .navbar-nav .open .dropdown-menu>.disabled>a:hover {
				color: #444;
				background-color: transparent;
			  }
			}
			
			@media screen and (min-width:768px) {
			  .jumbotron {
				padding-top: 48px;
				padding-bottom: 48px;
			  }
			  .container .jumbotron,
			  .container-fluid .jumbotron {
				padding-right: 60px;
				padding-left: 60px;
			  }
			  .jumbotron .h1,
			  .jumbotron h1 {
				font-size: 63px;
			  }
			}
			
			@media (min-width:768px) {
			  .modal-dialog {
				width: 600px;
				margin: 30px auto;
			  }
			  .modal-content {
				-webkit-box-shadow: 0 5px 15px rgba(0, 0, 0, .5);
				box-shadow: 0 5px 15px rgba(0, 0, 0, .5);
			  }
			  .modal-sm {
				width: 300px;
			  }
			}
			
			@media (min-width:992px) {
			  .modal-lg {
				width: 900px;
			  }
			}
			
			@media all and (transform-3d),
			(-webkit-transform-3d) {
			  .carousel-inner>.item {
				-webkit-transition: -webkit-transform .6s ease-in-out;
				-o-transition: -o-transform .6s ease-in-out;
				transition: transform .6s ease-in-out;
				-webkit-backface-visibility: hidden;
				backface-visibility: hidden;
				-webkit-perspective: 1000px;
				perspective: 1000px;
			  }
			  .carousel-inner>.item.active.right,
			  .carousel-inner>.item.next {
				left: 0;
				-webkit-transform: translate3d(100%, 0, 0);
				transform: translate3d(100%, 0, 0);
			  }
			  .carousel-inner>.item.active.left,
			  .carousel-inner>.item.prev {
				left: 0;
				-webkit-transform: translate3d(-100%, 0, 0);
				transform: translate3d(-100%, 0, 0);
			  }
			  .carousel-inner>.item.active,
			  .carousel-inner>.item.next.left,
			  .carousel-inner>.item.prev.right {
				left: 0;
				-webkit-transform: translate3d(0, 0, 0);
				transform: translate3d(0, 0, 0);
			  }
			}
			
			@media screen and (min-width:768px) {
			  .carousel-control .glyphicon-chevron-left,
			  .carousel-control .glyphicon-chevron-right,
			  .carousel-control .icon-next,
			  .carousel-control .icon-prev {
				width: 30px;
				height: 30px;
				margin-top: -10px;
				font-size: 30px;
			  }
			  .carousel-control .glyphicon-chevron-left,
			  .carousel-control .icon-prev {
				margin-left: -10px;
			  }
			  .carousel-control .glyphicon-chevron-right,
			  .carousel-control .icon-next {
				margin-right: -10px;
			  }
			  .carousel-caption {
				right: 20%;
				left: 20%;
				padding-bottom: 30px;
			  }
			  .carousel-indicators {
				bottom: 20px;
			  }
			}
			
			@media (max-width:767px) {
			  .visible-xs {
				display: block!important;
			  }
			  table.visible-xs {
				display: table!important;
			  }
			  tr.visible-xs {
				display: table-row!important;
			  }
			  td.visible-xs,
			  th.visible-xs {
				display: table-cell!important;
			  }
			}
			
			@media (max-width:767px) {
			  .visible-xs-block {
				display: block!important;
			  }
			}
			
			@media (max-width:767px) {
			  .visible-xs-inline {
				display: inline!important;
			  }
			}
			
			@media (max-width:767px) {
			  .visible-xs-inline-block {
				display: inline-block!important;
			  }
			}
			
			@media (min-width:768px) and (max-width:991px) {
			  .visible-sm {
				display: block!important;
			  }
			  table.visible-sm {
				display: table!important;
			  }
			  tr.visible-sm {
				display: table-row!important;
			  }
			  td.visible-sm,
			  th.visible-sm {
				display: table-cell!important;
			  }
			}
			
			@media (min-width:768px) and (max-width:991px) {
			  .visible-sm-block {
				display: block!important;
			  }
			}
			
			@media (min-width:768px) and (max-width:991px) {
			  .visible-sm-inline {
				display: inline!important;
			  }
			}
			
			@media (min-width:768px) and (max-width:991px) {
			  .visible-sm-inline-block {
				display: inline-block!important;
			  }
			}
			
			@media (min-width:992px) and (max-width:1199px) {
			  .visible-md {
				display: block!important;
			  }
			  table.visible-md {
				display: table!important;
			  }
			  tr.visible-md {
				display: table-row!important;
			  }
			  td.visible-md,
			  th.visible-md {
				display: table-cell!important;
			  }
			}
			
			@media (min-width:992px) and (max-width:1199px) {
			  .visible-md-block {
				display: block!important;
			  }
			}
			
			@media (min-width:992px) and (max-width:1199px) {
			  .visible-md-inline {
				display: inline!important;
			  }
			}
			
			@media (min-width:992px) and (max-width:1199px) {
			  .visible-md-inline-block {
				display: inline-block!important;
			  }
			}
			
			@media (min-width:1200px) {
			  .visible-lg {
				display: block!important;
			  }
			  table.visible-lg {
				display: table!important;
			  }
			  tr.visible-lg {
				display: table-row!important;
			  }
			  td.visible-lg,
			  th.visible-lg {
				display: table-cell!important;
			  }
			}
			
			@media (min-width:1200px) {
			  .visible-lg-block {
				display: block!important;
			  }
			}
			
			@media (min-width:1200px) {
			  .visible-lg-inline {
				display: inline!important;
			  }
			}
			
			@media (min-width:1200px) {
			  .visible-lg-inline-block {
				display: inline-block!important;
			  }
			}
			
			@media (max-width:767px) {
			  .hidden-xs {
				display: none!important;
			  }
			}
			
			@media (min-width:768px) and (max-width:991px) {
			  .hidden-sm {
				display: none!important;
			  }
			}
			
			@media (min-width:992px) and (max-width:1199px) {
			  .hidden-md {
				display: none!important;
			  }
			}
			
			@media (min-width:1200px) {
			  .hidden-lg {
				display: none!important;
			  }
			}
			
			@media print {
			  .visible-print {
				display: block!important;
			  }
			  table.visible-print {
				display: table!important;
			  }
			  tr.visible-print {
				display: table-row!important;
			  }
			  td.visible-print,
			  th.visible-print {
				display: table-cell!important;
			  }
			}
			
			@media print {
			  .visible-print-block {
				display: block!important;
			  }
			}
			
			@media print {
			  .visible-print-inline {
				display: inline!important;
			  }
			}
			
			@media print {
			  .visible-print-inline-block {
				display: inline-block!important;
			  }
			}
			
			@media print {
			  .hidden-print {
				display: none!important;
			  }
			}
		  </style>
		
		  <center class="wrapper" data-link-color="#1188E6" data-body-style="font-size: 14px; font-family: arial; color: #000000; background-color: #ebebeb;" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
			<div class="webkit" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-family: arial; font-size: 14px;">
			  <table cellpadding="0" cellspacing="0" border="0" width="100%" class="wrapper" bgcolor="#ebebeb" style="-moz-box-sizing: border-box; -moz-text-size-adjust: 100%; -ms-text-size-adjust: 100%; -webkit-box-sizing: border-box; -webkit-font-smoothing: antialiased; -webkit-text-size-adjust: 100%; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box; table-layout: fixed; width: 100% !important;">
				<tbody>
				  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
					<td valign="top" bgcolor="#ebebeb" width="100%" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; padding: 0;">
					  <table width="100%" role="content-container" class="outer" align="center" cellpadding="0" cellspacing="0" border="0" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box;">
						<tbody>
						  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
							<td width="100%" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; padding: 0;">
							  <table width="100%" cellpadding="0" cellspacing="0" border="0" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box;">
								<tbody>
								  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
									<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; padding: 0;">
									  <!--[if mso]>
								  <center>
								  <table><tr><td width="600">
								  <![endif]-->
									  <table width="100%" cellpadding="0" cellspacing="0" border="0" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box; max-width: 600px; width: 100%;"
										align="center">
										<tbody>
										  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
											<td role="modules-container" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; color: #000000; padding: 0px 0px 0px 0px; text-align: left;" bgcolor="#ffffff" width="100%" align="left">
		
											  <table class="module preheader preheader-hide" role="module" data-type="preheader" border="0" cellpadding="0" cellspacing="0" width="100%" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box; color: transparent; display: none !important; height: 0; mso-hide: all; opacity: 0; visibility: hidden; width: 0;">
												<tbody>
												  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
													<td role="module-content" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; padding: 0;">
													  <p style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-family: arial; font-size: 14px; margin: 0 0 10px; padding: 0;"></p>
													</td>
												  </tr>
												</tbody>
											  </table>
		
											  <table class="wrapper" role="module" data-type="image" border="0" cellpadding="0" cellspacing="0" width="100%" style="-moz-box-sizing: border-box; -moz-text-size-adjust: 100%; -ms-text-size-adjust: 100%; -webkit-box-sizing: border-box; -webkit-font-smoothing: antialiased; -webkit-text-size-adjust: 100%; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box; table-layout: fixed; width: 100% !important;">
												<tbody>
												  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
													<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-size: 6px; line-height: 10px; padding: 0px 0px 0px 0px;" valign="top" align="center">
													  <img class="max-width" border="0" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border: 0; box-sizing: border-box; color: #000000; display: block; font-family: Helvetica, arial, sans-serif; font-size: 16px; height: auto !important; max-width: 100% !important; text-decoration: none; vertical-align: middle; width: 100%;"
														src="https://marketing-image-production.s3.amazonaws.com/uploads/b09553df6918fe296ecfc240db655a184ff3a1317ee2d54e13d532e60177715894449dc25615184d98cf1c480c7935c7953cf6eeef5742bfdb6a66aecb5eabee.png" alt=""
														width="600">
													</td>
												  </tr>
												</tbody>
											  </table>
		
											  <table class="module" role="module" data-type="text" border="0" cellpadding="0" cellspacing="0" width="100%" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box; table-layout: fixed;">
												<tbody>
												  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
													<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; line-height: 22px; padding: 45px 45px 45px 45px; text-align: inherit;" height="100%" valign="top" bgcolor="">
													  <div style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-family: arial; font-size: 14px; text-align: center;"><strong style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-weight: 700;"><span style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; color: #3E3E3E;"><span style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-size: 20px;">Thanks For Contacting Me</span></span></strong></div>
													</td>
												  </tr>
												</tbody>
											  </table>
											  <table class="module" role="module" data-type="text" border="0" cellpadding="0" cellspacing="0" width="100%" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box; table-layout: fixed;">
												<tbody>
												  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
													<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; line-height: 22px; padding: 45px 45px 45px 45px; text-align: inherit;" height="100%" valign="top" bgcolor="">
													  <div style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-family: arial; font-size: 14px;"><strong style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-weight: 700;"><span style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; color: #3E3E3E;"><span style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-size: 20px;">Hello '. $receiverid.', <br style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">Thank you for contacting me, please allow me some time to look over your email. I understand you want a quick response so I will make it a top priorty to respond to you but while you\'re waiting check out my github and blog. Links are at the bottom of this email.  For reference here is a copy of your message. <br style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;"><br style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;"> Thank you, <br style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;"> Anthony Bible <br style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;"></span></span></strong></div>
													</td>
												  </tr>
												</tbody>
											  </table>
		
											  <table class="module" role="module" data-type="text" border="0" cellpadding="0" cellspacing="0" width="100%" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box; table-layout: fixed;">
												<tbody>
												  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
													<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; line-height: 22px; padding: 30px 45px 30px 45px; text-align: inherit;" height="100%" valign="top" bgcolor="">
													  <div style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-family: arial; font-size: 14px; text-align: center;"><span style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; color: #333333;"><table class="table table-striped table-dark" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box; margin-bottom: 20px; max-width: 100%; width: 100%;">
			<thead style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
			  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
				<th scope="col" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border-bottom: 2px solid #ddd; border-top: 0; box-sizing: border-box; line-height: 1.42857143; padding: 8px; text-align: left; vertical-align: bottom;">Entry</th>
				<th scope="col" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border-bottom: 2px solid #ddd; border-top: 0; box-sizing: border-box; line-height: 1.42857143; padding: 8px; text-align: left; vertical-align: bottom;">Input</th>
			  </tr>
			</thead>
			<tbody style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
			  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: #f9f9f9; box-sizing: border-box;">
				<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border-top: 1px solid #ddd; box-sizing: border-box; line-height: 1.42857143; padding: 8px; text-align: left; vertical-align: top;">Name</td>
				<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border-top: 1px solid #ddd; box-sizing: border-box; line-height: 1.42857143; padding: 8px; vertical-align: top;">'.$receiverid.'</td>
			  </tr>
			 
			  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: #f9f9f9; box-sizing: border-box;">
				<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border-top: 1px solid #ddd; box-sizing: border-box; line-height: 1.42857143; padding: 8px; text-align: left; vertical-align: top;">Phone</td>
				<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border-top: 1px solid #ddd; box-sizing: border-box; line-height: 1.42857143; padding: 8px; vertical-align: top;">'.$phone.'</td>
			  </tr>
			  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
				<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border-top: 1px solid #ddd; box-sizing: border-box; line-height: 1.42857143; padding: 8px; text-align: left; vertical-align: top;">Message</td>
				<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border-top: 1px solid #ddd; box-sizing: border-box; line-height: 1.42857143; padding: 8px; vertical-align: top;">
					'.$message .'
				</td>
			  </tr>
			</tbody>
		  </table>
		
		  
		
		</span></div>
													</td>
												  </tr>
		
												</tbody>
											  </table>
											  <table class="module" role="module" data-type="divider" border="0" cellpadding="0" cellspacing="0" width="100%" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box; table-layout: fixed;">
												<tbody>
												  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
													<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; padding: 0px 0px 0px 0px;" role="module-content" height="100%" valign="top" bgcolor="">
													  <table border="0" cellpadding="0" cellspacing="0" align="center" width="100%" height="5px" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box; font-size: 5px; line-height: 5px;">
														<tbody>
														  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
															<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; padding: 0px 0px 5px 0px;" bgcolor="#ebebeb"></td>
														  </tr>
														</tbody>
													  </table>
													</td>
												  </tr>
												</tbody>
											  </table>
											  <table class="module" role="module" data-type="social" align="center" border="0" cellpadding="0" cellspacing="0" width="100%" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box; table-layout: fixed;">
												<tbody style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
												  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
													<td valign="top" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: #F5F5F5; box-sizing: border-box; font-size: 6px; line-height: 10px; padding: 10px 0px 30px 0px;">
													  <table align="center" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; border-collapse: collapse; border-spacing: 0; box-sizing: border-box;">
														<tbody style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
														  <tr style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">
															<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; padding: 0px 5px;">
															  <a role="social-icon-link" href="facebook.com/bibleanthony1" target="_blank" alt="Facebook" data-nolink="false" title="Facebook " style="-moz-border-radius: 3px; -moz-box-sizing: border-box; -webkit-border-radius: 3px; -webkit-box-sizing: border-box; background-color: #3B579D; border-radius: 3px; box-sizing: border-box; color: #1188E6; display: inline-block; text-decoration: none;">
																<img role="social-icon" alt="Facebook" title="Facebook " height="30" width="30" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border: 0; box-sizing: border-box; height: 30px, width: 30px; vertical-align: middle;" src="https://marketing-image-production.s3.amazonaws.com/social/white/facebook.png">
															  </a>
															</td>
															<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; padding: 0px 5px;">
															  <a role="social-icon-link" href="twitter.com/_anthonybible" target="_blank" alt="Twitter" data-nolink="false" title="Twitter " style="-moz-border-radius: 3px; -moz-box-sizing: border-box; -webkit-border-radius: 3px; -webkit-box-sizing: border-box; background-color: #7AC4F7; border-radius: 3px; box-sizing: border-box; color: #1188E6; display: inline-block; text-decoration: none;">
																<img role="social-icon" alt="Twitter" title="Twitter " height="30" width="30" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border: 0; box-sizing: border-box; height: 30px, width: 30px; vertical-align: middle;" src="https://marketing-image-production.s3.amazonaws.com/social/white/twitter.png">
															  </a>
															</td>
															<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; padding: 0px 5px;">
															  <a role="social-icon-link" href="instagram.com/anthonybible" target="_blank" alt="Instagram" data-nolink="false" title="Instagram " style="-moz-border-radius: 3px; -moz-box-sizing: border-box; -webkit-border-radius: 3px; -webkit-box-sizing: border-box; background-color: #7F4B30; border-radius: 3px; box-sizing: border-box; color: #1188E6; display: inline-block; text-decoration: none;">
																<img role="social-icon" alt="Instagram" title="Instagram " height="30" width="30" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border: 0; box-sizing: border-box; height: 30px, width: 30px; vertical-align: middle;" src="https://marketing-image-production.s3.amazonaws.com/social/white/instagram.png">
															  </a>
															</td>
															<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; padding: 0px 5px;">
															  <a role="social-icon-link" href="https://www.linkedin.com/in/anthonybible/" target="_blank" alt="linkedin" data-nolink="false" title="linkedin" style="-moz-border-radius: 3px; -moz-box-sizing: border-box; -webkit-border-radius: 3px; -webkit-box-sizing: border-box; background-color: #0077B5; border-radius: 3px; box-sizing: border-box; color: #1188E6; display: inline-block; text-decoration: none;">
																<img role="social-icon" alt="Linkedin" title="linkedin" height="30" width="30" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border: 0; box-sizing: border-box; height: 30px, width: 30px; vertical-align: middle;" src="https://marketing-image-production.s3.amazonaws.com/social/white/linkedin.png">
															  </a>
															</td>
															<td style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; padding: 0px 5px;">
															  <a role="social-icon-link" href="https://github.com/Anthony-Bible" target="_blank" alt="github" data-nolink="false" title="github" style="-moz-border-radius: 3px; -moz-box-sizing: border-box; -webkit-border-radius: 3px; -webkit-box-sizing: border-box; background-color: #FFF; border-radius: 3px; box-sizing: border-box; color: #1188E6; display: inline-block; text-decoration: none;">
																<img role="social-icon" alt="github" title="github" height="30" width="30" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; border: 0; box-sizing: border-box; height: 30px, width: 30px; vertical-align: middle;" src="https://anthonybible.com/img/GitHub-Mark-64px.png">
															  </a>
															</td>
		
		
														  </tr>
														</tbody>
													  </table>
													</td>
		
		
		
		
												  </tr>
												</tbody>
											  </table>
											</td>
										  </tr>
										</tbody>
									  </table>
									</td>
								  </tr>
		
								</tbody>
							  </table>
		
		
							  <div data-role="module-unsubscribe" class="module unsubscribe-css__unsubscribe___2CDlR" role="module" data-type="unsubscribe" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: #ebebeb; box-sizing: border-box; color: #7a7a7a; font-family: arial; font-size: 11px; line-height: 20px; padding: 30px 0px 30px 0px; text-align: center;">
								<div class="Unsubscribe--addressLine" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-family: arial; font-size: 14px;">
								  <p class="Unsubscribe--senderName" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-family: ;font-size:11px; font-size: 14px; line-height: 20px; margin: 0 0 10px; padding: 0;">Anthony Bible</p>
								  <p style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-family: ;font-size:11px; font-size: 14px; line-height: 20px; margin: 0 0 10px; padding: 0;"><span class="Unsubscribe--senderAddress" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">PO Box 571442</span>, <span class="Unsubscribe--senderCity" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">Salt Lake City</span>,
									<span class="Unsubscribe--senderState" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">Utah</span> <span class="Unsubscribe--senderZip" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box;">84157</span>                            </p>
								</div>
								<p style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; box-sizing: border-box; font-family: ;font-size:11px; font-size: 14px; line-height: 20px; margin: 0 0 10px; padding: 0;"><a class="Unsubscribe--unsubscribeLink" href="&amp;lt;%asm_global_unsubscribe_raw_url%&amp;gt;" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; box-sizing: border-box; color: #2277ee; text-decoration: none;">Unsubscribe</a>                          - <a class="Unsubscribe--unsubscribePreferences" href="[Unsubscribe_Preferences]" style="-moz-box-sizing: border-box; -webkit-box-sizing: border-box; background-color: transparent; box-sizing: border-box; color: #2277ee; text-decoration: none;">Unsubscribe Preferences</a></p>
							  </div>
		
		
							</td>
						  </tr>
						</tbody>
					  </table>
					  <!--[if mso]>
								  </td></tr></table>
								  </center>
								  <![endif]-->
					</td>
				  </tr>
				</tbody>
			  </table>
		
		
		
		
		
		
			</div>
		  </center>
		
		</body>
		
		</html>';
		$char_set = 'UTF-8';
	

	try 
		{
			/* We've set all the parameters, it's now time to send it. To do this we just check the captcha response. If they failed we won't send the mail. This has dramatically reduced the spam to almost zero */
			
			$secret= getenv('GOOGLECAPTCHASECRET');
			$captchaResponse=$_POST["captcha"];
			echo "<response>";
			echo "<message>";

			$verifyUrl="https://www.google.com/recaptcha/api/siteverify?secret=$secret&response=$captchaResponse";
			$verify=file_get_contents($verifyUrl);
				$captcha_success=json_decode($verify);
				if ($captcha_success->success==false) {			
					echo "Looks like the robot overlords deterimined you were a bot, please try the Recaptcha again";
				}
				
				if ($captcha_success->success==true) {
				//This user is verified by recaptcha
				$result = $SesClient->sendEmail([
					'Destination' => [
						'ToAddresses' => $recieverEmail,
					],
					'ReplyToAddresses' => [$sender_email],
					'Source' => $sender_email,
					'Message' => [
					  'Body' => [
						  'Html' => [
							  'Charset' => $char_set,
							  'Data' => $html_body,
						  ],
						  'Text' => [
							  'Charset' => $char_set,
							  'Data' => $plaintext_body,
						  ],
					  ],
					  'Subject' => [
						  'Charset' => $char_set,
						  'Data' => $subject,
					  ],
					],
					
				]);
				$messageId = $result['MessageId'];
				echo "<h3>You Successfully sent the Email if you don't recieve an email please check your spam folder</h3>";



				}

			
			echo "</message>";
			echo "</response>";
			

		}
	 catch (Exception $e) {
	
			echo "<response>";
			echo "<message>";
			// output error message if fails
			echo $e->getMessage();
			echo "\n";
			echo "</message>";
			echo "</response>";
			
	}
}
sendEmail();










?>
